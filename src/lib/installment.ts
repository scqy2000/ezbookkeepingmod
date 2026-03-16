import {
    createEmptyInstallmentAccountRuleForm,
    createEmptyInstallmentPlanForm,
    InstallmentDueDateSources,
    InstallmentStorageModes,
    type InstallmentAccountRuleForm,
    type InstallmentAccountRuleInfoResponse,
    type InstallmentDistributionMode,
    type InstallmentItemRequest,
    type InstallmentPlanForm,
    type InstallmentPlanInfoResponse,
    type InstallmentPlanModifyRequest,
    type InstallmentPlanRequest,
} from "@/models/installment.ts";

function getMonthlyDateString(
    firstDueDate: string,
    monthlyDueDay: number,
    monthOffset: number,
): string {
    const date = new Date(`${firstDueDate}T00:00:00`);
    const targetMonth = new Date(
        date.getFullYear(),
        date.getMonth() + monthOffset,
        1,
    );
    const targetDay = Math.min(
        monthlyDueDay,
        new Date(
            targetMonth.getFullYear(),
            targetMonth.getMonth() + 1,
            0,
        ).getDate(),
    );
    targetMonth.setDate(targetDay);

    const year = targetMonth.getFullYear();
    const month = `${targetMonth.getMonth() + 1}`.padStart(2, "0");
    const day = `${targetMonth.getDate()}`.padStart(2, "0");

    return `${year}-${month}-${day}`;
}

function getClampedDateString(
    year: number,
    month: number,
    day: number,
): string {
    const targetMonth = new Date(year, month, 1);
    const targetDay = Math.min(
        day,
        new Date(
            targetMonth.getFullYear(),
            targetMonth.getMonth() + 1,
            0,
        ).getDate(),
    );
    targetMonth.setDate(targetDay);

    const targetYear = targetMonth.getFullYear();
    const targetMonthText = `${targetMonth.getMonth() + 1}`.padStart(2, "0");
    const targetDayText = `${targetMonth.getDate()}`.padStart(2, "0");

    return `${targetYear}-${targetMonthText}-${targetDayText}`;
}

function splitAmount(total: number, count: number): number[] {
    if (count < 1) {
        return [];
    }

    const base = Math.floor(total / count);
    const remainder = total - base * count;
    const amounts: number[] = [];

    for (let i = 0; i < count; i++) {
        amounts.push(base + (i < remainder ? 1 : 0));
    }

    return amounts;
}

export function generateInstallmentItems({
    principalTotal,
    feeTotal,
    periodCount,
    firstDueDate,
    monthlyDueDay,
    distributionMode,
    dueDates = [],
}: {
    principalTotal: number;
    feeTotal: number;
    periodCount: number;
    firstDueDate: string;
    monthlyDueDay: number;
    distributionMode: InstallmentDistributionMode;
    dueDates?: string[];
}): InstallmentItemRequest[] {
    if (
        !principalTotal ||
        periodCount < 1 ||
        (!firstDueDate && !dueDates.length)
    ) {
        return [];
    }

    const items: InstallmentItemRequest[] = [];
    const principalParts = splitAmount(principalTotal, periodCount);
    const feeParts = splitAmount(feeTotal, periodCount);

    if (distributionMode === "equal_total") {
        const totalParts = splitAmount(principalTotal + feeTotal, periodCount);

        for (let i = 0; i < periodCount; i++) {
            const feeAmount = feeParts[i] || 0;
            const dueAmount = totalParts[i] || 0;
            const principalAmount = Math.max(0, dueAmount - feeAmount);

            items.push({
                seqNo: i + 1,
                dueDate:
                    dueDates[i] ||
                    getMonthlyDateString(firstDueDate, monthlyDueDay, i),
                principalAmount: principalAmount,
                feeAmount: feeAmount,
                dueAmount: principalAmount + feeAmount,
            });
        }

        return items;
    }

    for (let i = 0; i < periodCount; i++) {
        const principalAmount = principalParts[i] || 0;
        const feeAmount = feeParts[i] || 0;

        items.push({
            seqNo: i + 1,
            dueDate:
                dueDates[i] ||
                getMonthlyDateString(firstDueDate, monthlyDueDay, i),
            principalAmount: principalAmount,
            feeAmount: feeAmount,
            dueAmount: principalAmount + feeAmount,
        });
    }

    return items;
}

export function normalizeInstallmentItems(
    items: InstallmentItemRequest[],
): InstallmentItemRequest[] {
    return items.map((item, index) => ({
        seqNo: index + 1,
        dueDate: item.dueDate,
        principalAmount: Math.max(0, Math.trunc(item.principalAmount || 0)),
        feeAmount: Math.max(0, Math.trunc(item.feeAmount || 0)),
        dueAmount: Math.max(
            0,
            Math.trunc((item.principalAmount || 0) + (item.feeAmount || 0)),
        ),
    }));
}

export function installmentFormFromResponse(
    plan: InstallmentPlanInfoResponse,
): InstallmentPlanForm {
    const form = createEmptyInstallmentPlanForm();
    const purchaseDate = new Date(plan.purchaseTime * 1000);

    form.id = plan.id;
    form.providerKey = plan.providerKey;
    form.customProviderName = plan.customProviderName || "";
    form.liabilityAccountId = plan.liabilityAccountId;
    form.defaultPaymentAccountId = plan.defaultPaymentAccountId || "";
    form.accountingMode = plan.accountingMode;
    form.purchaseTransactionId = plan.purchaseTransactionId || "";
    form.generatePurchaseTransaction = plan.generatedPurchaseTransaction;
    form.purchaseCategoryId = plan.purchaseCategoryId || "";
    form.repaymentCategoryId = plan.repaymentCategoryId || "";
    form.feeCategoryId = plan.feeCategoryId || "";
    form.transferCategoryId = plan.transferCategoryId || "";
    form.title = plan.title;
    form.notes = plan.notes;
    form.purchaseDatetime = new Date(
        purchaseDate.getTime() - purchaseDate.getTimezoneOffset() * 60000,
    )
        .toISOString()
        .slice(0, 16);
    form.currency = plan.currency;
    form.principalTotal = plan.principalTotal;
    form.feeTotal = plan.feeTotal;
    form.periodCount = plan.periodCount;
    form.dueDateSource = plan.dueDateSource;
    form.storageMode = plan.storageMode;
    form.accountRule = createEmptyInstallmentAccountRuleForm(
        plan.liabilityAccountId,
    );
    form.firstDueDate = plan.firstDueDate;
    form.monthlyDueDay = plan.monthlyDueDay;
    form.distributionMode = "custom";
    form.items = normalizeInstallmentItems(plan.items || []);

    return form;
}

export function installmentRequestFromForm(
    form: InstallmentPlanForm,
): InstallmentPlanRequest {
    const purchaseDate = new Date(form.purchaseDatetime);
    const items = normalizeInstallmentItems(form.items);
    const firstDueDate = items[0]?.dueDate || form.firstDueDate;
    const monthlyDueDay = Number(
        firstDueDate?.slice(-2) || form.monthlyDueDay || 1,
    );

    return {
        providerKey: form.providerKey,
        customProviderName: form.customProviderName,
        liabilityAccountId: form.liabilityAccountId,
        defaultPaymentAccountId: form.defaultPaymentAccountId || "0",
        accountingMode: Number(form.accountingMode) as 1 | 2,
        purchaseTransactionId: form.purchaseTransactionId || "0",
        generatePurchaseTransaction: form.generatePurchaseTransaction,
        purchaseCategoryId: form.purchaseCategoryId || "0",
        repaymentCategoryId: form.repaymentCategoryId || "0",
        feeCategoryId: form.feeCategoryId || "0",
        transferCategoryId: form.transferCategoryId || "0",
        title: form.title,
        notes: form.notes,
        purchaseTime: Math.floor(purchaseDate.getTime() / 1000),
        purchaseUtcOffset: -purchaseDate.getTimezoneOffset(),
        currency: form.currency,
        principalTotal: Math.trunc(form.principalTotal),
        feeTotal: Math.trunc(form.feeTotal),
        periodCount: Math.trunc(form.periodCount),
        dueDateSource: Number(
            form.dueDateSource || InstallmentDueDateSources.PlanRule,
        ) as 1 | 2,
        storageMode: Number(
            form.storageMode || InstallmentStorageModes.PlanItemsOnly,
        ) as 1 | 2,
        firstDueDate: firstDueDate,
        monthlyDueDay: Math.trunc(monthlyDueDay),
        items: items,
    };
}

export function installmentModifyRequestFromForm(
    form: InstallmentPlanForm,
): InstallmentPlanModifyRequest {
    return {
        id: form.id || "0",
        ...installmentRequestFromForm(form),
    };
}

export function getDefaultPayTimeValue(): string {
    const now = new Date();
    return new Date(now.getTime() - now.getTimezoneOffset() * 60000)
        .toISOString()
        .slice(0, 16);
}

export function getPayTimePayload(datetimeValue: string): {
    time: number;
    utcOffset: number;
} {
    const date = new Date(datetimeValue);

    return {
        time: Math.floor(date.getTime() / 1000),
        utcOffset: -date.getTimezoneOffset(),
    };
}

export function installmentAccountRuleFormFromResponse(
    rule: InstallmentAccountRuleInfoResponse | undefined,
    liabilityAccountId: string,
): InstallmentAccountRuleForm {
    if (!rule) {
        return createEmptyInstallmentAccountRuleForm(liabilityAccountId);
    }

    return {
        liabilityAccountId: rule.liabilityAccountId,
        statementDay: rule.statementDay,
        repaymentDay: rule.repaymentDay,
        timezone: rule.timezone,
        enabled: rule.enabled,
    };
}

export function generateInstallmentDueDatesByAccountRule({
    purchaseDatetime,
    periodCount,
    accountRule,
}: {
    purchaseDatetime: string;
    periodCount: number;
    accountRule: InstallmentAccountRuleForm;
}): string[] {
    if (!purchaseDatetime || periodCount < 1 || !accountRule.enabled) {
        return [];
    }

    const purchaseDate = new Date(purchaseDatetime);

    if (Number.isNaN(purchaseDate.getTime())) {
        return [];
    }

    const statementDay = Math.max(1, Math.trunc(accountRule.statementDay || 1));
    const repaymentDay = Math.max(1, Math.trunc(accountRule.repaymentDay || 1));
    const firstStatementDate = getClampedDateString(
        purchaseDate.getFullYear(),
        purchaseDate.getMonth(),
        statementDay,
    );
    const purchaseDateOnly = purchaseDatetime.slice(0, 10);
    let statementYear = purchaseDate.getFullYear();
    let statementMonth = purchaseDate.getMonth();

    if (purchaseDateOnly > firstStatementDate) {
        statementMonth += 1;

        if (statementMonth > 11) {
            statementMonth = 0;
            statementYear += 1;
        }
    }

    let firstDueMonth = statementMonth;
    let firstDueYear = statementYear;

    if (repaymentDay <= statementDay) {
        firstDueMonth += 1;

        if (firstDueMonth > 11) {
            firstDueMonth = 0;
            firstDueYear += 1;
        }
    }

    const dueDates: string[] = [];

    for (let i = 0; i < periodCount; i++) {
        const offsetMonth = firstDueMonth + i;
        const dueYear = firstDueYear + Math.floor(offsetMonth / 12);
        const dueMonth = offsetMonth % 12;
        dueDates.push(getClampedDateString(dueYear, dueMonth, repaymentDay));
    }

    return dueDates;
}
