<template>
    <v-row>
        <v-col cols="12">
            <v-card v-if="plan">
                <template #title>
                    <div class="d-flex align-center">
                        <span>{{ plan.title }}</span>
                        <v-spacer />
                        <v-btn class="me-2" color="default" variant="text" @click="router.push(`/installment/add?id=${plan.id}`)">{{ tt('Edit') }}</v-btn>
                        <v-btn color="default" variant="text" @click="router.back()">{{ tt('Back') }}</v-btn>
                    </div>
                </template>

                <v-alert class="mx-4 mt-2" type="error" variant="tonal" v-if="errorMessage">{{ errorMessage }}</v-alert>

                <v-card-text>
                    <div class="text-body-1 mb-4">{{ plan.providerName }} · {{ plan.currency }} · {{ plan.paidCount }}/{{ plan.periodCount }}</div>

                    <v-table>
                        <thead>
                        <tr>
                            <th>#</th>
                            <th>{{ tt('Due Date') }}</th>
                            <th>{{ tt('Principal') }}</th>
                            <th>{{ tt('Fee') }}</th>
                            <th>{{ tt('Total') }}</th>
                            <th>{{ tt('Status') }}</th>
                            <th>{{ tt('Operation') }}</th>
                        </tr>
                        </thead>
                        <tbody>
                        <tr :key="item.id" v-for="item in plan.items || []">
                            <td>{{ item.seqNo }}</td>
                            <td>{{ item.dueDate }}</td>
                            <td>{{ item.principalAmount }}</td>
                            <td>{{ item.feeAmount }}</td>
                            <td>{{ item.dueAmount }}</td>
                            <td>{{ item.status === InstallmentItemStatuses.Paid ? tt('Paid') : tt('Unpaid') }}</td>
                            <td>
                                <v-btn v-if="item.status !== InstallmentItemStatuses.Paid" density="comfortable" variant="text" color="primary" @click="openPayDialog(item)">{{ tt('Pay') }}</v-btn>
                                <v-btn v-else density="comfortable" variant="text" color="error" @click="unpay(item.id)">{{ tt('Unpay') }}</v-btn>
                            </td>
                        </tr>
                        </tbody>
                    </v-table>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>

    <v-dialog width="520" v-model="showPayDialog">
        <v-card>
            <template #title>{{ tt('Pay Installment') }}</template>
            <v-card-text>
                <v-select :items="paymentAccountOptions" item-title="name" item-value="id" v-model="payAccountId" :label="tt('Payment Account')"></v-select>
                <v-text-field type="datetime-local" v-model="payDatetime" :label="tt('Payment Time')"></v-text-field>
                <v-text-field v-model="payComment" :label="tt('Comment')"></v-text-field>
            </v-card-text>
            <v-card-actions>
                <v-spacer />
                <v-btn variant="text" @click="showPayDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="primary" :loading="paying" @click="pay">{{ tt('Confirm') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useI18n } from '@/locales/helpers.ts';

import { useAccountsStore } from '@/stores/account.ts';

import type { InstallmentItemInfoResponse, InstallmentPlanInfoResponse } from '@/models/installment.ts';
import { InstallmentItemStatuses } from '@/models/installment.ts';
import { getDefaultPayTimeValue, getPayTimePayload } from '@/lib/installment.ts';
import services from '@/lib/services.ts';

const { tt } = useI18n();
const route = useRoute();
const router = useRouter();
const accountsStore = useAccountsStore();

const plan = ref<InstallmentPlanInfoResponse | null>(null);
const errorMessage = ref<string>('');
const showPayDialog = ref<boolean>(false);
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
    const id = `${route.query['id'] || ''}`;

    if (!id) {
        errorMessage.value = 'Installment plan id is invalid';
        return;
    }

    Promise.all([
        accountsStore.loadAllAccounts({ force: false }),
        services.getInstallmentPlan({ id })
    ]).then(([, response]) => {
        const data = response.data;

        if (!data || !data.success || !data.result) {
            errorMessage.value = 'Unable to retrieve installment plan';
            return;
        }

        plan.value = data.result;
    }).catch(error => {
        errorMessage.value = error?.response?.data?.errorMessage || error?.message || 'Unable to retrieve installment plan';
    });
}

function openPayDialog(item: InstallmentItemInfoResponse): void {
    selectedItem.value = item;
    payAccountId.value = plan.value?.defaultPaymentAccountId || paymentAccountOptions.value[0]?.id || '';
    payDatetime.value = getDefaultPayTimeValue();
    payComment.value = '';
    showPayDialog.value = true;
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
        showPayDialog.value = false;
        paying.value = false;
        loadPlan();
    }).catch(error => {
        errorMessage.value = error?.response?.data?.errorMessage || error?.message || 'Unable to pay installment item';
        paying.value = false;
    });
}

function unpay(itemId: string): void {
    if (!window.confirm('Unpay this installment item?')) {
        return;
    }

    services.unpayInstallmentItem({ itemId }).then(() => {
        loadPlan();
    }).catch(error => {
        errorMessage.value = error?.response?.data?.errorMessage || error?.message || 'Unable to unpay installment item';
    });
}

onMounted(() => {
    loadPlan();
});
</script>
