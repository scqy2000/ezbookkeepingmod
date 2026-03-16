<template>
    <v-row>
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="d-flex align-center">
                        <span>{{
                            isEdit
                                ? tt("Edit Installment")
                                : tt("Add Installment")
                        }}</span>
                        <v-spacer />
                        <v-btn
                            color="default"
                            variant="text"
                            @click="router.back()"
                            >{{ tt("Back") }}</v-btn
                        >
                    </div>
                </template>

                <v-alert
                    class="mx-4 mt-2"
                    type="error"
                    variant="tonal"
                    v-if="errorMessage"
                    >{{ errorMessage }}</v-alert
                >

                <v-card-text>
                    <v-row>
                        <v-col cols="12" md="4">
                            <v-select
                                :items="providerOptions"
                                item-title="name"
                                item-value="key"
                                v-model="form.providerKey"
                                :label="tt('Provider')"
                            ></v-select>
                        </v-col>
                        <v-col
                            cols="12"
                            md="4"
                            v-if="form.providerKey === 'custom'"
                        >
                            <v-text-field
                                v-model="form.customProviderName"
                                :label="tt('Custom Provider Name')"
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-text-field
                                v-model="form.title"
                                :label="tt('Title')"
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="6">
                            <v-textarea
                                rows="3"
                                v-model="form.notes"
                                :label="tt('Notes')"
                            ></v-textarea>
                        </v-col>
                        <v-col cols="12" md="3">
                            <v-select
                                :items="liabilityAccountOptions"
                                item-title="name"
                                item-value="id"
                                v-model="form.liabilityAccountId"
                                :label="tt('Liability Account')"
                            ></v-select>
                        </v-col>
                        <v-col cols="12" md="3">
                            <v-select
                                :items="paymentAccountOptions"
                                item-title="name"
                                item-value="id"
                                v-model="form.defaultPaymentAccountId"
                                :label="tt('Default Payment Account')"
                            ></v-select>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-select
                                :items="accountingModeOptions"
                                item-title="name"
                                item-value="value"
                                v-model="form.accountingMode"
                                :label="tt('Accounting Mode')"
                                :disabled="hasLinkedPurchaseTransaction"
                            ></v-select>
                        </v-col>
                        <v-col cols="12" md="4" v-if="hasLinkedPurchaseTransaction">
                            <v-text-field
                                v-model="form.purchaseTransactionId"
                                :label="tt('Linked Purchase Transaction')"
                                readonly
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-text-field
                                type="datetime-local"
                                v-model="form.purchaseDatetime"
                                :label="tt('Purchase Time')"
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-text-field
                                v-model="form.currency"
                                :label="tt('Currency')"
                                readonly
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-text-field
                                type="number"
                                v-model.number="form.principalTotal"
                                :label="tt('Principal Total')"
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-text-field
                                type="number"
                                v-model.number="form.feeTotal"
                                :label="tt('Fee Total')"
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-text-field
                                type="number"
                                v-model.number="form.periodCount"
                                :label="tt('Installment Count')"
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-select
                                :items="dueDateSourceOptions"
                                item-title="name"
                                item-value="value"
                                v-model="form.dueDateSource"
                                :label="tt('Due Date Source')"
                            ></v-select>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-select
                                :items="storageModeOptions"
                                item-title="name"
                                item-value="value"
                                v-model="form.storageMode"
                                :label="tt('Future Repayment Storage')"
                            ></v-select>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-text-field
                                type="date"
                                v-model="form.firstDueDate"
                                :label="tt('First Due Date')"
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-text-field
                                type="number"
                                v-model.number="form.monthlyDueDay"
                                :label="tt('Monthly Due Day')"
                            ></v-text-field>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-select
                                :items="distributionModeOptions"
                                item-title="name"
                                item-value="value"
                                v-model="form.distributionMode"
                                :label="tt('Generation Mode')"
                            ></v-select>
                        </v-col>
                        <v-col
                            cols="12"
                            md="4"
                            v-if="
                                form.accountingMode ===
                                InstallmentAccountingModes.PurchaseRecognized
                            "
                        >
                            <v-switch
                                color="primary"
                                v-model="form.generatePurchaseTransaction"
                                :label="tt('Generate Purchase Transaction')"
                                :disabled="hasLinkedPurchaseTransaction"
                            ></v-switch>
                        </v-col>
                        <v-col
                            cols="12"
                            md="4"
                            v-if="
                                form.accountingMode ===
                                    InstallmentAccountingModes.PurchaseRecognized &&
                                form.generatePurchaseTransaction
                            "
                        >
                            <v-select
                                :items="expenseCategoryOptions"
                                item-title="name"
                                item-value="id"
                                v-model="form.purchaseCategoryId"
                                :label="tt('Purchase Category')"
                            ></v-select>
                        </v-col>
                        <v-col
                            cols="12"
                            md="4"
                            v-if="
                                form.accountingMode ===
                                InstallmentAccountingModes.PurchaseRecognized
                            "
                        >
                            <v-select
                                :items="expenseCategoryOptions"
                                item-title="name"
                                item-value="id"
                                v-model="form.feeCategoryId"
                                :label="tt('Fee Category')"
                            ></v-select>
                        </v-col>
                        <v-col
                            cols="12"
                            md="4"
                            v-if="
                                form.accountingMode ===
                                InstallmentAccountingModes.RepaymentRecognized
                            "
                        >
                            <v-select
                                :items="expenseCategoryOptions"
                                item-title="name"
                                item-value="id"
                                v-model="form.repaymentCategoryId"
                                :label="tt('Repayment Category')"
                            ></v-select>
                        </v-col>
                        <v-col cols="12" md="4">
                            <v-select
                                :items="transferCategoryOptions"
                                item-title="name"
                                item-value="id"
                                v-model="form.transferCategoryId"
                                :label="tt('Transfer Category')"
                            ></v-select>
                        </v-col>
                        <template
                            v-if="
                                form.dueDateSource ===
                                InstallmentDueDateSources.AccountRule
                            "
                        >
                            <v-col cols="12" md="3">
                                <v-text-field
                                    type="number"
                                    v-model.number="
                                        form.accountRule.statementDay
                                    "
                                    :label="tt('Statement Day')"
                                ></v-text-field>
                            </v-col>
                            <v-col cols="12" md="3">
                                <v-text-field
                                    type="number"
                                    v-model.number="
                                        form.accountRule.repaymentDay
                                    "
                                    :label="tt('Repayment Day')"
                                ></v-text-field>
                            </v-col>
                            <v-col cols="12" md="4">
                                <v-text-field
                                    v-model="form.accountRule.timezone"
                                    :label="tt('Rule Timezone')"
                                ></v-text-field>
                            </v-col>
                            <v-col cols="12" md="2">
                                <v-switch
                                    color="primary"
                                    v-model="form.accountRule.enabled"
                                    :label="tt('Rule Enabled')"
                                ></v-switch>
                            </v-col>
                        </template>
                    </v-row>

                    <div class="d-flex align-center mt-4 mb-2">
                        <span class="text-subtitle-1">{{
                            tt("Installment Items")
                        }}</span>
                        <v-btn
                            class="ms-3"
                            color="default"
                            variant="outlined"
                            @click="regenerateItems"
                            >{{ tt("Regenerate") }}</v-btn
                        >
                        <v-btn
                            class="ms-2"
                            color="default"
                            variant="text"
                            @click="addItem"
                            >{{ tt("Add Item") }}</v-btn
                        >
                    </div>

                    <v-table>
                        <thead>
                            <tr>
                                <th>#</th>
                                <th>{{ tt("Due Date") }}</th>
                                <th>{{ tt("Principal") }}</th>
                                <th>{{ tt("Fee") }}</th>
                                <th>{{ tt("Total") }}</th>
                                <th>{{ tt("Operation") }}</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr
                                :key="index"
                                v-for="(item, index) in form.items"
                            >
                                <td>{{ index + 1 }}</td>
                                <td>
                                    <v-text-field
                                        density="compact"
                                        hide-details
                                        type="date"
                                        v-model="item.dueDate"
                                    ></v-text-field>
                                </td>
                                <td>
                                    <v-text-field
                                        density="compact"
                                        hide-details
                                        type="number"
                                        v-model.number="item.principalAmount"
                                    ></v-text-field>
                                </td>
                                <td>
                                    <v-text-field
                                        density="compact"
                                        hide-details
                                        type="number"
                                        v-model.number="item.feeAmount"
                                    ></v-text-field>
                                </td>
                                <td>
                                    {{
                                        (item.principalAmount || 0) +
                                        (item.feeAmount || 0)
                                    }}
                                </td>
                                <td>
                                    <v-btn
                                        density="comfortable"
                                        variant="text"
                                        color="error"
                                        @click="removeItem(index)"
                                        >{{ tt("Delete") }}</v-btn
                                    >
                                </td>
                            </tr>
                        </tbody>
                    </v-table>

                    <div class="d-flex justify-end mt-4">
                        <v-btn
                            color="primary"
                            :loading="submitting"
                            @click="save"
                            >{{ tt("Save") }}</v-btn
                        >
                    </div>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";

import { useI18n } from "@/locales/helpers.ts";

import { useAccountsStore } from "@/stores/account.ts";
import { useTransactionCategoriesStore } from "@/stores/transactionCategory.ts";

import { CategoryType } from "@/core/category.ts";
import type { TransactionCategory } from "@/models/transaction_category.ts";
import type { Account } from "@/models/account.ts";
import {
    createEmptyInstallmentAccountRuleForm,
    createEmptyInstallmentPlanForm,
    InstallmentAccountingModes,
    InstallmentDueDateSources,
    InstallmentProviders,
    InstallmentStorageModes,
    type InstallmentPlanForm,
} from "@/models/installment.ts";
import {
    applyInstallmentPrefillFromQuery,
    generateInstallmentDueDatesByAccountRule,
    generateInstallmentItems,
    installmentFormFromResponse,
    installmentModifyRequestFromForm,
    installmentRequestFromForm,
    normalizeInstallmentItems,
} from "@/lib/installment.ts";
import services from "@/lib/services.ts";

const { tt } = useI18n();
const route = useRoute();
const router = useRouter();
const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();

const loading = ref<boolean>(true);
const submitting = ref<boolean>(false);
const errorMessage = ref<string>("");
const form = reactive<InstallmentPlanForm>(createEmptyInstallmentPlanForm());

const isEdit = computed<boolean>(() => !!route.query["id"]);
const hasLinkedPurchaseTransaction = computed<boolean>(
    () => !!form.purchaseTransactionId && !form.generatePurchaseTransaction,
);
const providerOptions = computed(() => InstallmentProviders);
const accountingModeOptions = computed(() => [
    {
        name: tt("Purchase Recognized"),
        value: InstallmentAccountingModes.PurchaseRecognized,
    },
    {
        name: tt("Repayment Recognized"),
        value: InstallmentAccountingModes.RepaymentRecognized,
    },
]);
const dueDateSourceOptions = computed(() => [
    { name: tt("Plan Rule"), value: InstallmentDueDateSources.PlanRule },
    { name: tt("Account Rule"), value: InstallmentDueDateSources.AccountRule },
]);
const storageModeOptions = computed(() => [
    {
        name: tt("Plan Items Only"),
        value: InstallmentStorageModes.PlanItemsOnly,
    },
    {
        name: tt("Generated Scheduled Templates"),
        value: InstallmentStorageModes.GeneratedScheduleTemplates,
    },
]);
const distributionModeOptions = computed(() => [
    { name: tt("Equal Total"), value: "equal_total" },
    { name: tt("Equal Principal And Fee"), value: "equal_principal_fee" },
    { name: tt("Custom"), value: "custom" },
]);

const liabilityAccountOptions = computed<Account[]>(() =>
    accountsStore.allVisiblePlainAccounts.filter(
        (account) => account.isLiability,
    ),
);
const paymentAccountOptions = computed<Account[]>(() =>
    accountsStore.allVisiblePlainAccounts.filter(
        (account) => account.isAsset && account.currency === form.currency,
    ),
);
const expenseCategoryOptions = computed(() =>
    flattenSubCategories(
        transactionCategoriesStore.allTransactionCategories[
            CategoryType.Expense
        ] || [],
    ),
);
const transferCategoryOptions = computed(() =>
    flattenSubCategories(
        transactionCategoriesStore.allTransactionCategories[
            CategoryType.Transfer
        ] || [],
    ),
);

function flattenSubCategories(
    categories: TransactionCategory[],
): { id: string; name: string }[] {
    const result: { id: string; name: string }[] = [];

    for (const category of categories) {
        for (const subCategory of category.subCategories || []) {
            if (subCategory.visible) {
                result.push({
                    id: subCategory.id,
                    name: `${category.name} / ${subCategory.name}`,
                });
            }
        }
    }

    return result;
}

function assignForm(nextForm: InstallmentPlanForm): void {
    Object.assign(form, nextForm);
}

function regenerateItems(): void {
    const dueDates =
        form.dueDateSource === InstallmentDueDateSources.AccountRule
            ? generateInstallmentDueDatesByAccountRule({
                  purchaseDatetime: form.purchaseDatetime,
                  periodCount: form.periodCount,
                  accountRule: form.accountRule,
              })
            : [];

    if (dueDates.length > 0) {
        const firstDueDate = dueDates[0]!;
        form.firstDueDate = firstDueDate;
        form.monthlyDueDay = Number(firstDueDate.slice(-2));
    }

    form.items = generateInstallmentItems({
        principalTotal: form.principalTotal,
        feeTotal: form.feeTotal,
        periodCount: form.periodCount,
        firstDueDate: form.firstDueDate,
        monthlyDueDay: form.monthlyDueDay,
        distributionMode: form.distributionMode,
        dueDates: dueDates,
    });
}

function addItem(): void {
    form.items = normalizeInstallmentItems([
        ...(form.items || []),
        {
            seqNo: form.items.length + 1,
            dueDate: form.firstDueDate,
            principalAmount: 0,
            feeAmount: 0,
            dueAmount: 0,
        },
    ]);
    form.distributionMode = "custom";
}

function removeItem(index: number): void {
    form.items = normalizeInstallmentItems(
        form.items.filter((_, itemIndex) => itemIndex !== index),
    );
    form.periodCount = form.items.length;
}

function loadInitialData(): void {
    loading.value = true;

    Promise.all([
        accountsStore.loadAllAccounts({ force: false }),
        transactionCategoriesStore.loadAllCategories({ force: false }),
    ])
        .then(() => {
            const planId = route.query["id"];

            if (!planId) {
                applyInstallmentPrefillFromQuery(form, route.query);
                regenerateItems();
                loading.value = false;
                return;
            }

            services
                .getInstallmentPlan({ id: `${planId}` })
                .then((response) => {
                    const data = response.data;

                    if (!data || !data.success || !data.result) {
                        errorMessage.value =
                            "Unable to retrieve installment plan";
                        loading.value = false;
                        return;
                    }

                    assignForm(installmentFormFromResponse(data.result));
                    loadAccountRule(form.liabilityAccountId);
                    loading.value = false;
                })
                .catch((error) => {
                    errorMessage.value =
                        error?.response?.data?.errorMessage ||
                        error?.message ||
                        "Unable to retrieve installment plan";
                    loading.value = false;
                });
        })
        .catch((error) => {
            errorMessage.value =
                error?.response?.data?.errorMessage ||
                error?.message ||
                "Unable to load page data";
            loading.value = false;
        });
}

function loadAccountRule(liabilityAccountId: string): void {
    if (!liabilityAccountId || liabilityAccountId === "0") {
        Object.assign(
            form.accountRule,
            createEmptyInstallmentAccountRuleForm(liabilityAccountId || ""),
        );
        return;
    }

    services
        .getInstallmentAccountRule({ liabilityAccountId: liabilityAccountId })
        .then((response) => {
            const data = response.data;

            if (!data || !data.success || !data.result) {
                return;
            }

            form.accountRule.liabilityAccountId =
                data.result.liabilityAccountId;
            form.accountRule.statementDay = data.result.statementDay;
            form.accountRule.repaymentDay = data.result.repaymentDay;
            form.accountRule.timezone = data.result.timezone;
            form.accountRule.enabled = data.result.enabled;
        })
        .catch(() => {
            Object.assign(
                form.accountRule,
                createEmptyInstallmentAccountRuleForm(liabilityAccountId),
            );
        });
}

function save(): void {
    submitting.value = true;
    errorMessage.value = "";

    const savePlan = () => {
        const promise = isEdit.value
            ? services.modifyInstallmentPlan(
                  installmentModifyRequestFromForm(form),
              )
            : services.addInstallmentPlan(installmentRequestFromForm(form));

        promise
            .then((response) => {
                const data = response.data;

                if (!data || !data.success || !data.result) {
                    errorMessage.value = isEdit.value
                        ? "Unable to save installment plan"
                        : "Unable to add installment plan";
                    submitting.value = false;
                    return;
                }

                submitting.value = false;
                router.push(`/installment/detail?id=${data.result.id}`);
            })
            .catch((error) => {
                errorMessage.value =
                    error?.response?.data?.errorMessage ||
                    error?.message ||
                    "Unable to save installment plan";
                submitting.value = false;
            });
    };

    if (form.dueDateSource === InstallmentDueDateSources.AccountRule) {
        services
            .saveInstallmentAccountRule({
                liabilityAccountId: form.liabilityAccountId,
                statementDay: Number(form.accountRule.statementDay),
                repaymentDay: Number(form.accountRule.repaymentDay),
                timezone: form.accountRule.timezone,
                enabled: form.accountRule.enabled,
            })
            .then(() => {
                savePlan();
            })
            .catch((error) => {
                errorMessage.value =
                    error?.response?.data?.errorMessage ||
                    error?.message ||
                    "Unable to save installment account rule";
                submitting.value = false;
            });
        return;
    }

    savePlan();
}

watch(
    () => form.firstDueDate,
    (value) => {
        if (
            form.dueDateSource === InstallmentDueDateSources.PlanRule &&
            value &&
            value.length === 10
        ) {
            form.monthlyDueDay = parseInt(value.slice(-2));
        }
    },
);

watch(
    () => form.liabilityAccountId,
    (value) => {
        const account = liabilityAccountOptions.value.find(
            (item) => item.id === value,
        );

        if (account) {
            form.currency = account.currency;
        }

        form.accountRule.liabilityAccountId = value || "";
        loadAccountRule(value || "");
    },
);

watch(
    () => [
        form.principalTotal,
        form.feeTotal,
        form.periodCount,
        form.firstDueDate,
        form.monthlyDueDay,
        form.distributionMode,
        form.dueDateSource,
        form.purchaseDatetime,
        form.accountRule.statementDay,
        form.accountRule.repaymentDay,
        form.accountRule.enabled,
    ],
    () => {
        if (form.distributionMode !== "custom") {
            regenerateItems();
        }
    },
);

onMounted(() => {
    loadInitialData();
});
</script>
