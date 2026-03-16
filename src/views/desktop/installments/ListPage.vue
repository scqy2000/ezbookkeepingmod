<template>
    <v-row>
        <v-col cols="12">
            <v-card>
                <template #title>
                    <div class="d-flex align-center">
                        <span>{{ tt('Installments') }}</span>
                        <v-btn class="ms-3" color="primary" variant="tonal" @click="router.push('/installment/add')">{{ tt('Add') }}</v-btn>
                        <v-btn class="ms-2" color="default" variant="text" :disabled="loading" @click="loadPlans">{{ tt('Refresh') }}</v-btn>
                    </div>
                </template>

                <v-alert class="mx-4 mt-2" type="error" variant="tonal" v-if="errorMessage">{{ errorMessage }}</v-alert>

                <v-table>
                    <thead>
                    <tr>
                        <th>{{ tt('Title') }}</th>
                        <th>{{ tt('Type') }}</th>
                        <th>{{ tt('Provider') }}</th>
                        <th>{{ tt('Currency') }}</th>
                        <th>{{ tt('Installments') }}</th>
                        <th>{{ tt('Next Due Date') }}</th>
                        <th>{{ tt('Operation') }}</th>
                    </tr>
                    </thead>
                    <tbody>
                    <tr v-if="loading">
                        <td colspan="7">{{ tt('Loading') }}...</td>
                    </tr>
                    <tr v-else-if="!plans.length">
                        <td colspan="7">{{ tt('No available installment plans') }}</td>
                    </tr>
                    <tr :key="plan.id" v-for="plan in plans">
                        <td>{{ plan.title }}</td>
                        <td>{{ plan.accountingMode === InstallmentAccountingModes.PurchaseRecognized ? tt('Purchase Recognized') : tt('Repayment Recognized') }}</td>
                        <td>{{ plan.providerName }}</td>
                        <td>{{ plan.currency }}</td>
                        <td>{{ plan.paidCount }}/{{ plan.periodCount }}</td>
                        <td>{{ plan.nextUnpaidItem?.dueDate || '-' }}</td>
                        <td>
                            <v-btn class="me-2" density="comfortable" variant="text" @click="router.push(`/installment/detail?id=${plan.id}`)">{{ tt('View') }}</v-btn>
                            <v-btn class="me-2" density="comfortable" variant="text" @click="router.push(`/installment/add?id=${plan.id}`)">{{ tt('Edit') }}</v-btn>
                            <v-btn density="comfortable" variant="text" color="error" @click="remove(plan.id)">{{ tt('Delete') }}</v-btn>
                        </td>
                    </tr>
                    </tbody>
                </v-table>
            </v-card>
        </v-col>
    </v-row>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useI18n } from '@/locales/helpers.ts';

import services from '@/lib/services.ts';
import type { InstallmentPlanInfoResponse } from '@/models/installment.ts';
import { InstallmentAccountingModes } from '@/models/installment.ts';

const { tt } = useI18n();
const router = useRouter();

const loading = ref<boolean>(true);
const errorMessage = ref<string>('');
const plans = ref<InstallmentPlanInfoResponse[]>([]);

function loadPlans(): void {
    loading.value = true;
    errorMessage.value = '';

    services.getAllInstallmentPlans().then(response => {
        const data = response.data;

        if (!data || !data.success || !data.result) {
            errorMessage.value = 'Unable to retrieve installment plans';
            loading.value = false;
            return;
        }

        plans.value = data.result;
        loading.value = false;
    }).catch(error => {
        errorMessage.value = error?.response?.data?.errorMessage || error?.message || 'Unable to retrieve installment plans';
        loading.value = false;
    });
}

function remove(id: string): void {
    if (!window.confirm('Delete this installment plan?')) {
        return;
    }

    services.deleteInstallmentPlan({ id }).then(() => {
        loadPlans();
    }).catch(error => {
        errorMessage.value = error?.response?.data?.errorMessage || error?.message || 'Unable to delete installment plan';
    });
}

onMounted(() => {
    loadPlans();
});
</script>
