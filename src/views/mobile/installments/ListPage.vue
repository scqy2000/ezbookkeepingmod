<template>
    <f7-page ptr @ptr:refresh="reload">
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt('Installments')"></f7-nav-title>
            <f7-nav-right>
                <f7-link icon-f7="plus" href="/installment/add"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers v-if="loading">
            <f7-list-item title="Loading..."></f7-list-item>
        </f7-list>

        <f7-block strong inset v-else-if="errorMessage">{{ errorMessage }}</f7-block>

        <f7-list strong inset dividers v-else-if="plans.length">
            <f7-list-item :key="plan.id" :title="plan.title" :footer="`${plan.providerName} · ${plan.currency}`" :after="`${plan.paidCount}/${plan.periodCount}`" :link="`/installment/detail?id=${plan.id}`" v-for="plan in plans">
                <template #text>
                    <div>{{ plan.nextUnpaidItem?.dueDate || '-' }}</div>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers v-else>
            <f7-list-item :title="tt('No available installment plans')"></f7-list-item>
        </f7-list>
    </f7-page>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { Router } from 'framework7/types';

import { useI18n } from '@/locales/helpers.ts';
import type { InstallmentPlanInfoResponse } from '@/models/installment.ts';
import services from '@/lib/services.ts';

defineProps<{
    f7route: Router.Route;
    f7router: Router.Router;
}>();

const { tt } = useI18n();
const loading = ref<boolean>(true);
const errorMessage = ref<string>('');
const plans = ref<InstallmentPlanInfoResponse[]>([]);

function loadPlans(done?: () => void): void {
    loading.value = !done;
    errorMessage.value = '';

    services.getAllInstallmentPlans().then(response => {
        const data = response.data;

        if (!data || !data.success || !data.result) {
            errorMessage.value = 'Unable to retrieve installment plans';
            loading.value = false;
            done?.();
            return;
        }

        plans.value = data.result;
        loading.value = false;
        done?.();
    }).catch(error => {
        errorMessage.value = error?.response?.data?.errorMessage || error?.message || 'Unable to retrieve installment plans';
        loading.value = false;
        done?.();
    });
}

function reload(done: () => void): void {
    loadPlans(done);
}

onMounted(() => {
    loadPlans();
});
</script>
