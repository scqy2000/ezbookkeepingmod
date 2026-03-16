<template>
    <f7-page>
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="plan?.title || tt('Installment Detail')"></f7-nav-title>
            <f7-nav-right v-if="plan">
                <f7-link icon-f7="pencil" :href="`/installment/add?id=${plan.id}`"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-block strong inset v-if="errorMessage">{{ errorMessage }}</f7-block>

        <f7-card v-if="plan">
            <f7-card-content>
                <div>{{ plan.providerName }} · {{ plan.currency }}</div>
                <div>{{ plan.paidCount }}/{{ plan.periodCount }}</div>
            </f7-card-content>
        </f7-card>

        <f7-list strong inset dividers v-if="plan">
            <f7-list-item :key="item.id" :title="`#${item.seqNo} · ${item.dueDate}`" :after="`${item.dueAmount}`" v-for="item in plan.items || []">
                <template #text>
                    <div>{{ item.status === InstallmentItemStatuses.Paid ? tt('Paid') : tt('Unpaid') }}</div>
                </template>
                <template #footer>
                    <f7-link color="blue" @click="openPay(item)" v-if="item.status !== InstallmentItemStatuses.Paid">{{ tt('Pay') }}</f7-link>
                    <f7-link color="red" @click="unpay(item.id)" v-else>{{ tt('Unpay') }}</f7-link>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-popup swipe-to-close v-model:opened="showPayPopup">
            <f7-page>
                <f7-navbar>
                    <f7-nav-left>
                        <f7-link popup-close>{{ tt('Cancel') }}</f7-link>
                    </f7-nav-left>
                    <f7-nav-title :title="tt('Pay Installment')"></f7-nav-title>
                </f7-navbar>

                <f7-list strong inset dividers>
                    <f7-list-input type="select" :label="tt('Payment Account')" v-model:value="payAccountId">
                        <option :key="account.id" :value="account.id" v-for="account in paymentAccountOptions">{{ account.name }}</option>
                    </f7-list-input>
                    <f7-list-input type="datetime-local" :label="tt('Payment Time')" v-model:value="payDatetime"></f7-list-input>
                    <f7-list-input type="text" :label="tt('Comment')" v-model:value="payComment"></f7-list-input>
                </f7-list>

                <f7-block>
                    <f7-button fill large :loading="paying" @click="pay">{{ tt('Confirm') }}</f7-button>
                </f7-block>
            </f7-page>
        </f7-popup>
    </f7-page>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';

import { useAccountsStore } from '@/stores/account.ts';

import type { InstallmentItemInfoResponse, InstallmentPlanInfoResponse } from '@/models/installment.ts';
import { InstallmentItemStatuses } from '@/models/installment.ts';
import { getDefaultPayTimeValue, getPayTimePayload } from '@/lib/installment.ts';
import services from '@/lib/services.ts';

const props = defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const accountsStore = useAccountsStore();

const plan = ref<InstallmentPlanInfoResponse | null>(null);
const errorMessage = ref<string>('');
const showPayPopup = ref<boolean>(false);
const selectedItem = ref<InstallmentItemInfoResponse | null>(null);
const payAccountId = ref<string>('');
const payDatetime = ref<string>(getDefaultPayTimeValue());
const payComment = ref<string>('');
const paying = ref<boolean>(false);

const paymentAccountOptions = computed(() => accountsStore.allVisiblePlainAccounts.filter(account => account.isAsset && account.currency === (plan.value?.currency || '')).map(account => ({
    id: account.id,
    name: account.name
})));

function loadPlan(): void {
    const id = `${props.f7route.query['id'] || ''}`;

    if (!id) {
        errorMessage.value = tt('Installment plan id is invalid');
        return;
    }

    Promise.all([
        accountsStore.loadAllAccounts({ force: false }),
        services.getInstallmentPlan({ id })
    ]).then(([, response]) => {
        const data = response.data;

        if (!data || !data.success || !data.result) {
            errorMessage.value = tt('Unable to retrieve installment plan');
            return;
        }

        plan.value = data.result;
    }).catch(error => {
        errorMessage.value = error?.response?.data?.errorMessage || error?.message || tt('Unable to retrieve installment plan');
    });
}

function openPay(item: InstallmentItemInfoResponse): void {
    selectedItem.value = item;
    payAccountId.value = plan.value?.defaultPaymentAccountId || paymentAccountOptions.value[0]?.id || '';
    payDatetime.value = getDefaultPayTimeValue();
    payComment.value = '';
    showPayPopup.value = true;
}

function pay(): void {
    if (!selectedItem.value) {
        return;
    }

    paying.value = true;
    const payload = getPayTimePayload(payDatetime.value);

    services.payInstallmentItem({
        itemId: selectedItem.value.id,
        paymentAccountId: payAccountId.value,
        time: payload.time,
        utcOffset: payload.utcOffset,
        comment: payComment.value
    }).then(() => {
        paying.value = false;
        showPayPopup.value = false;
        loadPlan();
    }).catch(error => {
        errorMessage.value = error?.response?.data?.errorMessage || error?.message || tt('Unable to pay installment item');
        paying.value = false;
    });
}

function unpay(itemId: string): void {
    services.unpayInstallmentItem({ itemId }).then(() => {
        loadPlan();
    }).catch(error => {
        errorMessage.value = error?.response?.data?.errorMessage || error?.message || tt('Unable to unpay installment item');
    });
}

onMounted(() => {
    loadPlan();
});
</script>
