export type InstallmentAccountingMode = 1 | 2;
export type InstallmentDueDateSource = 1 | 2;
export type InstallmentStorageMode = 1 | 2;
export type InstallmentItemStatus = 1 | 2;
export type InstallmentDistributionMode =
    | "equal_total"
    | "equal_principal_fee"
    | "custom";

export const InstallmentAccountingModes = {
    PurchaseRecognized: 1 as InstallmentAccountingMode,
    RepaymentRecognized: 2 as InstallmentAccountingMode,
};

export const InstallmentDueDateSources = {
    PlanRule: 1 as InstallmentDueDateSource,
    AccountRule: 2 as InstallmentDueDateSource,
};

export const InstallmentStorageModes = {
    PlanItemsOnly: 1 as InstallmentStorageMode,
    GeneratedScheduleTemplates: 2 as InstallmentStorageMode,
};

export const InstallmentItemStatuses = {
    Unpaid: 1 as InstallmentItemStatus,
    Paid: 2 as InstallmentItemStatus,
};

export const InstallmentProviders = [
    { key: "huabei", name: "花呗" },
    { key: "jd_baitiao", name: "京东白条" },
    { key: "custom", name: "Custom" },
];

export interface InstallmentItemRequest {
    seqNo: number;
    dueDate: string;
    principalAmount: number;
    feeAmount: number;
    dueAmount: number;
}

export interface InstallmentAccountRuleSaveRequest {
    liabilityAccountId: string;
    statementDay: number;
    repaymentDay: number;
    timezone: string;
    enabled: boolean;
}

export interface InstallmentAccountRuleDeleteRequest {
    liabilityAccountId: string;
}

export interface InstallmentAccountRuleInfoResponse {
    id: string;
    liabilityAccountId: string;
    statementDay: number;
    repaymentDay: number;
    timezone: string;
    enabled: boolean;
}

export interface InstallmentAccountRuleForm {
    liabilityAccountId: string;
    statementDay: number;
    repaymentDay: number;
    timezone: string;
    enabled: boolean;
}

export interface InstallmentPlanRequest {
    providerKey: string;
    customProviderName: string;
    liabilityAccountId: string;
    defaultPaymentAccountId: string;
    accountingMode: InstallmentAccountingMode;
    purchaseTransactionId: string;
    generatePurchaseTransaction: boolean;
    purchaseCategoryId: string;
    repaymentCategoryId: string;
    feeCategoryId: string;
    transferCategoryId: string;
    title: string;
    notes: string;
    purchaseTime: number;
    purchaseUtcOffset: number;
    currency: string;
    principalTotal: number;
    feeTotal: number;
    periodCount: number;
    dueDateSource: InstallmentDueDateSource;
    storageMode: InstallmentStorageMode;
    firstDueDate: string;
    monthlyDueDay: number;
    items: InstallmentItemRequest[];
}

export interface InstallmentPlanModifyRequest extends InstallmentPlanRequest {
    id: string;
}

export interface InstallmentPlanDeleteRequest {
    id: string;
}

export interface InstallmentItemPayRequest {
    itemId: string;
    paymentAccountId: string;
    time: number;
    utcOffset: number;
    comment: string;
    transferCategoryId?: string;
    repaymentCategoryId?: string;
    feeCategoryId?: string;
}

export interface InstallmentItemUnpayRequest {
    itemId: string;
}

export interface InstallmentItemInfoResponse {
    id: string;
    seqNo: number;
    dueDate: string;
    principalAmount: number;
    feeAmount: number;
    dueAmount: number;
    status: InstallmentItemStatus;
    paidTime: number;
    expenseTransactionId?: string;
    repaymentTransactionId?: string;
    feeTransactionId?: string;
    generatedTemplateId?: string;
}

export interface InstallmentPlanInfoResponse {
    id: string;
    providerKey: string;
    providerName: string;
    customProviderName?: string;
    liabilityAccountId: string;
    defaultPaymentAccountId?: string;
    accountingMode: InstallmentAccountingMode;
    purchaseTransactionId?: string;
    generatedPurchaseTransaction: boolean;
    purchaseCategoryId?: string;
    repaymentCategoryId?: string;
    feeCategoryId?: string;
    transferCategoryId?: string;
    title: string;
    notes: string;
    purchaseTime: number;
    purchaseUtcOffset: number;
    currency: string;
    principalTotal: number;
    feeTotal: number;
    dueTotal: number;
    periodCount: number;
    dueDateSource: InstallmentDueDateSource;
    storageMode: InstallmentStorageMode;
    firstDueDate: string;
    monthlyDueDay: number;
    paidCount: number;
    unpaidCount: number;
    overdueCount: number;
    nextUnpaidItem?: InstallmentItemInfoResponse;
    items?: InstallmentItemInfoResponse[];
}

export interface InstallmentPlanForm {
    id?: string;
    providerKey: string;
    customProviderName: string;
    liabilityAccountId: string;
    defaultPaymentAccountId: string;
    accountingMode: InstallmentAccountingMode;
    purchaseTransactionId: string;
    generatePurchaseTransaction: boolean;
    purchaseCategoryId: string;
    repaymentCategoryId: string;
    feeCategoryId: string;
    transferCategoryId: string;
    title: string;
    notes: string;
    purchaseDatetime: string;
    currency: string;
    principalTotal: number;
    feeTotal: number;
    periodCount: number;
    dueDateSource: InstallmentDueDateSource;
    storageMode: InstallmentStorageMode;
    accountRule: InstallmentAccountRuleForm;
    firstDueDate: string;
    monthlyDueDay: number;
    distributionMode: InstallmentDistributionMode;
    items: InstallmentItemRequest[];
}

export function createEmptyInstallmentAccountRuleForm(
    liabilityAccountId = "",
): InstallmentAccountRuleForm {
    return {
        liabilityAccountId: liabilityAccountId,
        statementDay: 1,
        repaymentDay: 10,
        timezone:
            Intl.DateTimeFormat().resolvedOptions().timeZone || "Asia/Shanghai",
        enabled: true,
    };
}

export function createEmptyInstallmentPlanForm(): InstallmentPlanForm {
    const now = new Date();
    const purchaseDatetime = new Date(
        now.getTime() - now.getTimezoneOffset() * 60000,
    )
        .toISOString()
        .slice(0, 16);
    const firstDueDate = new Date(
        now.getTime() - now.getTimezoneOffset() * 60000,
    )
        .toISOString()
        .slice(0, 10);

    return {
        providerKey: InstallmentProviders[0]!.key,
        customProviderName: "",
        liabilityAccountId: "",
        defaultPaymentAccountId: "",
        accountingMode: InstallmentAccountingModes.PurchaseRecognized,
        purchaseTransactionId: "",
        generatePurchaseTransaction: true,
        purchaseCategoryId: "",
        repaymentCategoryId: "",
        feeCategoryId: "",
        transferCategoryId: "",
        title: "",
        notes: "",
        purchaseDatetime: purchaseDatetime,
        currency: "CNY",
        principalTotal: 0,
        feeTotal: 0,
        periodCount: 3,
        dueDateSource: InstallmentDueDateSources.PlanRule,
        storageMode: InstallmentStorageModes.PlanItemsOnly,
        accountRule: createEmptyInstallmentAccountRuleForm(),
        firstDueDate: firstDueDate,
        monthlyDueDay: parseInt(firstDueDate.slice(-2)),
        distributionMode: "equal_principal_fee",
        items: [],
    };
}

export function getInstallmentProviderDisplayName(
    providerKey: string,
    customProviderName?: string,
): string {
    const provider = InstallmentProviders.find(
        (item) => item.key === providerKey,
    );

    if (!provider) {
        return providerKey;
    }

    if (providerKey === "custom") {
        return customProviderName || provider.name;
    }

    return provider.name;
}
