package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
)

func TestGetInstallmentProviderName(t *testing.T) {
	assert.Equal(t, "花呗", GetInstallmentProviderName(INSTALLMENT_PROVIDER_HUABEI, ""))
	assert.Equal(t, "京东白条", GetInstallmentProviderName(INSTALLMENT_PROVIDER_JD_BAITIAO, ""))
	assert.Equal(t, "自定义", GetInstallmentProviderName(INSTALLMENT_PROVIDER_CUSTOM, "自定义"))
}

func TestValidateInstallmentItems(t *testing.T) {
	items := []*InstallmentItem{
		{
			SeqNo:           1,
			DueDate:         "2026-04-09",
			PrincipalAmount: 3000,
			FeeAmount:       100,
			DueAmount:       3100,
		},
		{
			SeqNo:           2,
			DueDate:         "2026-05-09",
			PrincipalAmount: 3000,
			FeeAmount:       100,
			DueAmount:       3100,
		},
	}

	assert.Nil(t, ValidateInstallmentItems(items, 6000, 200))
}

func TestValidateInstallmentItems_InvalidAmount(t *testing.T) {
	items := []*InstallmentItem{
		{
			SeqNo:           1,
			DueDate:         "2026-04-09",
			PrincipalAmount: 3000,
			FeeAmount:       100,
			DueAmount:       3000,
		},
	}

	assert.ErrorIs(t, ValidateInstallmentItems(items, 3000, 100), errs.ErrInstallmentAmountMismatch)
}

func TestInstallmentPlanToResponse(t *testing.T) {
	plan := &InstallmentPlan{
		PlanId:                    1001,
		ProviderKey:               INSTALLMENT_PROVIDER_HUABEI,
		Title:                     "Phone",
		PrincipalTotal:            9000,
		FeeTotal:                  300,
		PeriodCount:               3,
		DueDateSource:             INSTALLMENT_DUE_DATE_SOURCE_PLAN_RULE,
		StorageMode:               INSTALLMENT_STORAGE_MODE_PLAN_ITEMS_ONLY,
		FirstDueDate:              "2026-04-09",
		MonthlyDueDay:             9,
		PurchaseTime:              100,
		Currency:                  "CNY",
		AccountingMode:            INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED,
		PurchaseTimezoneUtcOffset: 480,
	}
	items := []*InstallmentItem{
		{
			ItemId:          1,
			SeqNo:           1,
			DueDate:         "2026-04-09",
			PrincipalAmount: 3000,
			FeeAmount:       100,
			DueAmount:       3100,
			Status:          INSTALLMENT_ITEM_STATUS_UNPAID,
		},
		{
			ItemId:          2,
			SeqNo:           2,
			DueDate:         "2026-05-09",
			PrincipalAmount: 3000,
			FeeAmount:       100,
			DueAmount:       3100,
			Status:          INSTALLMENT_ITEM_STATUS_PAID,
		},
	}

	resp := plan.ToInstallmentPlanInfoResponse(items, true)

	assert.Equal(t, int64(9300), resp.DueTotal)
	assert.Equal(t, int16(1), resp.PaidCount)
	assert.Equal(t, int16(1), resp.UnpaidCount)
	assert.NotNil(t, resp.NextUnpaidItem)
	assert.Equal(t, int16(1), resp.NextUnpaidItem.SeqNo)
	assert.Len(t, resp.Items, 2)
}
