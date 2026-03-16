<template>
    <f7-page>
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title
                :title="isEdit ? tt('Edit Installment') : tt('Add Installment')"
            ></f7-nav-title>
        </f7-navbar>

        <f7-block strong inset v-if="errorMessage">{{ errorMessage }}</f7-block>

        <f7-list strong inset dividers>
            <f7-list-input
                type="text"
                :label="tt('Title')"
                v-model:value="form.title"
            ></f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Provider')"
                v-model:value="form.providerKey"
            >
                <option
                    :key="provider.key"
                    :value="provider.key"
                    v-for="provider in providerOptions"
                >
                    {{ provider.name }}
                </option>
            </f7-list-input>
            <f7-list-input
                type="text"
                :label="tt('Custom Provider Name')"
                v-model:value="form.customProviderName"
                v-if="form.providerKey === 'custom'"
            ></f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Liability Account')"
                v-model:value="form.liabilityAccountId"
            >
                <option
                    :key="account.id"
                    :value="account.id"
                    v-for="account in liabilityAccountOptions"
                >
                    {{ account.name }}
                </option>
            </f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Default Payment Account')"
                v-model:value="form.defaultPaymentAccountId"
            >
                <option value="0">-</option>
                <option
                    :key="account.id"
                    :value="account.id"
                    v-for="account in paymentAccountOptions"
                >
                    {{ account.name }}
                </option>
            </f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Accounting Mode')"
                v-model:value="form.accountingMode"
                :disabled="hasLinkedPurchaseTransaction"
            >
                <option :value="InstallmentAccountingModes.PurchaseRecognized">
                    {{ tt('Purchase Recognized') }}
                </option>
                <option :value="InstallmentAccountingModes.RepaymentRecognized">
                    {{ tt('Repayment Recognized') }}
                </option>
            </f7-list-input>
            <f7-list-input
                type="text"
                :label="tt('Linked Purchase Transaction')"
                v-model:value="form.purchaseTransactionId"
                readonly
                v-if="hasLinkedPurchaseTransaction"
            ></f7-list-input>
            <f7-list-input
                type="datetime-local"
                :label="tt('Purchase Time')"
                v-model:value="form.purchaseDatetime"
            ></f7-list-input>
            <f7-list-input
                type="number"
                :label="tt('Principal Total')"
                v-model:value="form.principalTotal"
            ></f7-list-input>
            <f7-list-input
                type="number"
                :label="tt('Fee Total')"
                v-model:value="form.feeTotal"
            ></f7-list-input>
            <f7-list-input
                type="number"
                :label="tt('Installment Count')"
                v-model:value="form.periodCount"
            ></f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Due Date Source')"
                v-model:value="form.dueDateSource"
            >
                <option :value="InstallmentDueDateSources.PlanRule">
                    {{ tt('Plan Rule') }}
                </option>
                <option :value="InstallmentDueDateSources.AccountRule">
                    {{ tt('Account Rule') }}
                </option>
            </f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Future Repayment Storage')"
                v-model:value="form.storageMode"
            >
                <option :value="InstallmentStorageModes.PlanItemsOnly">
                    {{ tt('Plan Items Only') }}
                </option>
                <option
                    :value="InstallmentStorageModes.GeneratedScheduleTemplates"
                >
                    {{ tt('Generated Scheduled Templates') }}
                </option>
            </f7-list-input>
            <f7-list-input
                type="date"
                :label="tt('First Due Date')"
                v-model:value="form.firstDueDate"
            ></f7-list-input>
            <f7-list-input
                type="number"
                :label="tt('Monthly Due Day')"
                v-model:value="form.monthlyDueDay"
            ></f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Generation Mode')"
                v-model:value="form.distributionMode"
            >
                <option value="equal_total">{{ tt('Equal Total') }}</option>
                <option value="equal_principal_fee">
                    {{ tt('Equal Principal And Fee') }}
                </option>
                <option value="custom">{{ tt('Custom') }}</option>
            </f7-list-input>
            <f7-list-item
                v-if="
                    Number(form.accountingMode) ===
                    InstallmentAccountingModes.PurchaseRecognized
                "
            >
                <span>{{ tt("Generate Purchase Transaction") }}</span>
                <f7-toggle
                    :checked="form.generatePurchaseTransaction"
                    :disabled="hasLinkedPurchaseTransaction"
                    @toggle:change="form.generatePurchaseTransaction = $event"
                ></f7-toggle>
            </f7-list-item>
            <f7-list-input
                type="select"
                :label="tt('Purchase Category')"
                v-model:value="form.purchaseCategoryId"
                v-if="
                    Number(form.accountingMode) ===
                        InstallmentAccountingModes.PurchaseRecognized &&
                    form.generatePurchaseTransaction
                "
            >
                <option
                    :key="item.id"
                    :value="item.id"
                    v-for="item in expenseCategoryOptions"
                >
                    {{ item.name }}
                </option>
            </f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Fee Category')"
                v-model:value="form.feeCategoryId"
                v-if="
                    Number(form.accountingMode) ===
                    InstallmentAccountingModes.PurchaseRecognized
                "
            >
                <option
                    :key="item.id"
                    :value="item.id"
                    v-for="item in expenseCategoryOptions"
                >
                    {{ item.name }}
                </option>
            </f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Repayment Category')"
                v-model:value="form.repaymentCategoryId"
                v-if="
                    Number(form.accountingMode) ===
                    InstallmentAccountingModes.RepaymentRecognized
                "
            >
                <option
                    :key="item.id"
                    :value="item.id"
                    v-for="item in expenseCategoryOptions"
                >
                    {{ item.name }}
                </option>
            </f7-list-input>
            <f7-list-input
                type="select"
                :label="tt('Transfer Category')"
                v-model:value="form.transferCategoryId"
            >
                <option
                    :key="item.id"
                    :value="item.id"
                    v-for="item in transferCategoryOptions"
                >
                    {{ item.name }}
                </option>
            </f7-list-input>
            <template
                v-if="
                    Number(form.dueDateSource) ===
                    InstallmentDueDateSources.AccountRule
                "
            >
                <f7-list-input
                    type="number"
                    :label="tt('Statement Day')"
                    v-model:value="form.accountRule.statementDay"
                ></f7-list-input>
                <f7-list-input
                    type="number"
                    :label="tt('Repayment Day')"
                    v-model:value="form.accountRule.repaymentDay"
                ></f7-list-input>
                <f7-list-input
                    type="text"
                    :label="tt('Rule Timezone')"
                    v-model:value="form.accountRule.timezone"
                ></f7-list-input>
                <f7-list-item>
                    <span>{{ tt("Rule Enabled") }}</span>
                    <f7-toggle
                        :checked="form.accountRule.enabled"
                        @toggle:change="form.accountRule.enabled = $event"
                    ></f7-toggle>
                </f7-list-item>
            </template>
        </f7-list>

        <f7-block class="display-flex justify-content-space-between">
            <f7-button fill tonal @click="regenerateItems">{{
                tt("Regenerate")
            }}</f7-button>
            <f7-button fill tonal @click="addItem">{{
                tt("Add Item")
            }}</f7-button>
        </f7-block>

        <f7-list strong inset dividers>
            <f7-list-item :key="index" v-for="(item, index) in form.items">
                <f7-list-input
                    type="date"
                    :label="`#${index + 1} ${tt('Due Date')}`"
                    v-model:value="item.dueDate"
                ></f7-list-input>
                <f7-list-input
                    type="number"
                    :label="tt('Principal')"
                    v-model:value="item.principalAmount"
                ></f7-list-input>
                <f7-list-input
                    type="number"
                    :label="tt('Fee')"
                    v-model:value="item.feeAmount"
                ></f7-list-input>
                <f7-link color="red" @click="removeItem(index)">{{
                    tt("Delete")
                }}</f7-link>
            </f7-list-item>
        </f7-list>

        <f7-block>
            <f7-button fill large :loading="submitting" @click="save">{{
                tt("Save")
            }}</f7-button>
        </f7-block>
    </f7-page>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import type { Router } from "framework7/types";

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

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const accountsStore = useAccountsStore();
const transactionCategoriesStore = useTransactionCategoriesStore();

const form = reactive<InstallmentPlanForm>(createEmptyInstallmentPlanForm());
const errorMessage = ref<string>("");
const submitting = ref<boolean>(false);
const isEdit = computed<boolean>(() => !!props.f7route.query["id"]);
const hasLinkedPurchaseTransaction = computed<boolean>(
    () => !!form.purchaseTransactionId && !form.generatePurchaseTransaction,
);
const providerOptions = computed(() =>
    InstallmentProviders.map((provider) => ({
        ...provider,
        name: provider.key === "custom" ? tt("Custom") : provider.name,
    })),
);
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
        Number(form.dueDateSource) === InstallmentDueDateSources.AccountRule
            ? generateInstallmentDueDatesByAccountRule({
                  purchaseDatetime: form.purchaseDatetime,
                  periodCount: Number(form.periodCount),
                  accountRule: form.accountRule,
              })
            : [];

    if (dueDates.length > 0) {
        const firstDueDate = dueDates[0]!;
        form.firstDueDate = firstDueDate;
        form.monthlyDueDay = Number(firstDueDate.slice(-2));
    }

    form.items = generateInstallmentItems({
        principalTotal: Number(form.principalTotal),
        feeTotal: Number(form.feeTotal),
        periodCount: Number(form.periodCount),
        firstDueDate: form.firstDueDate,
        monthlyDueDay: Number(form.monthlyDueDay),
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

function save(): void {
    submitting.value = true;
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
                    errorMessage.value = tt("Unable to save installment plan");
                    submitting.value = false;
                    return;
                }

                submitting.value = false;
                props.f7router.navigate(
                    `/installment/detail?id=${data.result.id}`,
                );
            })
            .catch((error) => {
                errorMessage.value =
                    error?.response?.data?.errorMessage ||
                    error?.message ||
                    tt("Unable to save installment plan");
                submitting.value = false;
            });
    };

    if (Number(form.dueDateSource) === InstallmentDueDateSources.AccountRule) {
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
                    tt("Unable to save installment account rule");
                submitting.value = false;
            });
        return;
    }

    savePlan();
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

onMounted(() => {
    Promise.all([
        accountsStore.loadAllAccounts({ force: false }),
        transactionCategoriesStore.loadAllCategories({ force: false }),
    ])
        .then(() => {
            const id = props.f7route.query["id"];

            if (!id) {
                applyInstallmentPrefillFromQuery(form, props.f7route.query);
                regenerateItems();
                return;
            }

            services
                .getInstallmentPlan({ id: `${id}` })
                .then((response) => {
                    const data = response.data;

                    if (!data || !data.success || !data.result) {
                        errorMessage.value =
                            tt("Unable to retrieve installment plan");
                        return;
                    }

                    assignForm(installmentFormFromResponse(data.result));
                    loadAccountRule(form.liabilityAccountId);
                })
                .catch((error) => {
                    errorMessage.value =
                        error?.response?.data?.errorMessage ||
                        error?.message ||
                        tt("Unable to retrieve installment plan");
                });
        })
        .catch((error) => {
            errorMessage.value =
                error?.response?.data?.errorMessage ||
                error?.message ||
                tt("Unable to load page data");
        });
});

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
    () => form.firstDueDate,
    (value) => {
        if (
            Number(form.dueDateSource) === InstallmentDueDateSources.PlanRule &&
            value &&
            value.length === 10
        ) {
            form.monthlyDueDay = parseInt(value.slice(-2));
        }
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
</script>
