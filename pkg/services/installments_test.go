package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

const installmentTestUid int64 = 1001

func TestInstallmentService_CreatePlanAndPayItem_PurchaseRecognized(t *testing.T) {
	ctx, cleanup := prepareInstallmentServiceTestEnv(t)
	defer cleanup()

	seedInstallmentTestBaseData(t, ctx)

	plan := &models.InstallmentPlan{
		Uid:                          installmentTestUid,
		ProviderKey:                  models.INSTALLMENT_PROVIDER_HUABEI,
		LiabilityAccountId:           2001,
		DefaultPaymentAccountId:      2002,
		AccountingMode:               models.INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED,
		GeneratedPurchaseTransaction: true,
		PurchaseCategoryId:           3002,
		FeeCategoryId:                3003,
		TransferCategoryId:           4002,
		Title:                        "Phone",
		PurchaseTime:                 1710000000,
		PurchaseTimezoneUtcOffset:    480,
		Currency:                     "CNY",
		PrincipalTotal:               9000,
		FeeTotal:                     300,
		PeriodCount:                  3,
		DueDateSource:                models.INSTALLMENT_DUE_DATE_SOURCE_PLAN_RULE,
		StorageMode:                  models.INSTALLMENT_STORAGE_MODE_PLAN_ITEMS_ONLY,
		FirstDueDate:                 "2026-04-09",
		MonthlyDueDay:                9,
	}
	items := []*models.InstallmentItem{
		{SeqNo: 1, DueDate: "2026-04-09", PrincipalAmount: 3000, FeeAmount: 100, DueAmount: 3100},
		{SeqNo: 2, DueDate: "2026-05-09", PrincipalAmount: 3000, FeeAmount: 100, DueAmount: 3100},
		{SeqNo: 3, DueDate: "2026-06-09", PrincipalAmount: 3000, FeeAmount: 100, DueAmount: 3100},
	}

	err := Installments.CreatePlan(ctx, plan, items)
	assert.Nil(t, err)
	assert.NotZero(t, plan.PlanId)
	assert.NotZero(t, plan.PurchaseTransactionId)

	allItems, err := Installments.GetItemsByPlanId(ctx, installmentTestUid, plan.PlanId)
	assert.Nil(t, err)
	assert.Len(t, allItems, 3)

	err = Installments.PayItem(ctx, installmentTestUid, &models.InstallmentItemPayRequest{
		ItemId:    allItems[0].ItemId,
		Time:      1710000100,
		UtcOffset: 480,
	})
	assert.Nil(t, err)

	paidItems, err := Installments.GetItemsByPlanId(ctx, installmentTestUid, plan.PlanId)
	assert.Nil(t, err)
	assert.Equal(t, models.INSTALLMENT_ITEM_STATUS_PAID, paidItems[0].Status)
	assert.NotZero(t, paidItems[0].RepaymentTransactionId)
	assert.NotZero(t, paidItems[0].FeeTransactionId)

	accounts, err := Accounts.GetAccountsByAccountIds(ctx, installmentTestUid, []int64{2001, 2002})
	assert.Nil(t, err)
	assert.Equal(t, int64(-6000), accounts[2001].Balance)
	assert.Equal(t, int64(96900), accounts[2002].Balance)

	err = Installments.UnpayItem(ctx, installmentTestUid, paidItems[0].ItemId)
	assert.Nil(t, err)

	unpaidItems, err := Installments.GetItemsByPlanId(ctx, installmentTestUid, plan.PlanId)
	assert.Nil(t, err)
	assert.Equal(t, models.INSTALLMENT_ITEM_STATUS_UNPAID, unpaidItems[0].Status)

	accounts, err = Accounts.GetAccountsByAccountIds(ctx, installmentTestUid, []int64{2001, 2002})
	assert.Nil(t, err)
	assert.Equal(t, int64(-9000), accounts[2001].Balance)
	assert.Equal(t, int64(100000), accounts[2002].Balance)
}

func TestInstallmentService_PayItem_RepaymentRecognized(t *testing.T) {
	ctx, cleanup := prepareInstallmentServiceTestEnv(t)
	defer cleanup()

	seedInstallmentTestBaseData(t, ctx)

	plan := &models.InstallmentPlan{
		Uid:                       installmentTestUid,
		ProviderKey:               models.INSTALLMENT_PROVIDER_JD_BAITIAO,
		LiabilityAccountId:        2001,
		DefaultPaymentAccountId:   2002,
		AccountingMode:            models.INSTALLMENT_ACCOUNTING_MODE_REPAYMENT_RECOGNIZED,
		RepaymentCategoryId:       3002,
		TransferCategoryId:        4002,
		Title:                     "Laptop",
		PurchaseTime:              1710000000,
		PurchaseTimezoneUtcOffset: 480,
		Currency:                  "CNY",
		PrincipalTotal:            6000,
		FeeTotal:                  0,
		PeriodCount:               2,
		DueDateSource:             models.INSTALLMENT_DUE_DATE_SOURCE_PLAN_RULE,
		StorageMode:               models.INSTALLMENT_STORAGE_MODE_PLAN_ITEMS_ONLY,
		FirstDueDate:              "2026-04-15",
		MonthlyDueDay:             15,
	}
	items := []*models.InstallmentItem{
		{SeqNo: 1, DueDate: "2026-04-15", PrincipalAmount: 3000, FeeAmount: 0, DueAmount: 3000},
		{SeqNo: 2, DueDate: "2026-05-15", PrincipalAmount: 3000, FeeAmount: 0, DueAmount: 3000},
	}

	err := Installments.CreatePlan(ctx, plan, items)
	assert.Nil(t, err)

	allItems, err := Installments.GetItemsByPlanId(ctx, installmentTestUid, plan.PlanId)
	assert.Nil(t, err)

	err = Installments.PayItem(ctx, installmentTestUid, &models.InstallmentItemPayRequest{
		ItemId:    allItems[0].ItemId,
		Time:      1710000200,
		UtcOffset: 480,
	})
	assert.Nil(t, err)

	accounts, err := Accounts.GetAccountsByAccountIds(ctx, installmentTestUid, []int64{2001, 2002})
	assert.Nil(t, err)
	assert.Equal(t, int64(0), accounts[2001].Balance)
	assert.Equal(t, int64(97000), accounts[2002].Balance)
}

func TestInstallmentService_DeletePlanRejectsPaidPlan(t *testing.T) {
	ctx, cleanup := prepareInstallmentServiceTestEnv(t)
	defer cleanup()

	seedInstallmentTestBaseData(t, ctx)

	plan := &models.InstallmentPlan{
		Uid:                          installmentTestUid,
		ProviderKey:                  models.INSTALLMENT_PROVIDER_HUABEI,
		LiabilityAccountId:           2001,
		DefaultPaymentAccountId:      2002,
		AccountingMode:               models.INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED,
		GeneratedPurchaseTransaction: true,
		PurchaseCategoryId:           3002,
		TransferCategoryId:           4002,
		Title:                        "Phone",
		PurchaseTime:                 1710000000,
		PurchaseTimezoneUtcOffset:    480,
		Currency:                     "CNY",
		PrincipalTotal:               3000,
		FeeTotal:                     0,
		PeriodCount:                  1,
		DueDateSource:                models.INSTALLMENT_DUE_DATE_SOURCE_PLAN_RULE,
		StorageMode:                  models.INSTALLMENT_STORAGE_MODE_PLAN_ITEMS_ONLY,
		FirstDueDate:                 "2026-04-09",
		MonthlyDueDay:                9,
	}
	items := []*models.InstallmentItem{
		{SeqNo: 1, DueDate: "2026-04-09", PrincipalAmount: 3000, FeeAmount: 0, DueAmount: 3000},
	}

	err := Installments.CreatePlan(ctx, plan, items)
	assert.Nil(t, err)

	allItems, err := Installments.GetItemsByPlanId(ctx, installmentTestUid, plan.PlanId)
	assert.Nil(t, err)

	err = Installments.PayItem(ctx, installmentTestUid, &models.InstallmentItemPayRequest{
		ItemId:    allItems[0].ItemId,
		Time:      1710000100,
		UtcOffset: 480,
	})
	assert.Nil(t, err)

	err = Installments.DeletePlan(ctx, installmentTestUid, plan.PlanId)
	assert.ErrorIs(t, err, errs.ErrInstallmentHasPaidItemsCannotDelete)
}

func TestInstallmentService_SaveAccountRuleAndUseAccountRulePlan(t *testing.T) {
	ctx, cleanup := prepareInstallmentServiceTestEnv(t)
	defer cleanup()

	seedInstallmentTestBaseData(t, ctx)

	rule := &models.InstallmentAccountRule{
		Uid:                installmentTestUid,
		LiabilityAccountId: 2001,
		StatementDay:       5,
		RepaymentDay:       15,
		Timezone:           "Asia/Shanghai",
		Enabled:            true,
	}

	err := Installments.SaveAccountRule(ctx, rule)
	assert.Nil(t, err)
	assert.NotZero(t, rule.RuleId)

	savedRule, err := Installments.GetAccountRuleByLiabilityAccountId(ctx, installmentTestUid, 2001)
	assert.Nil(t, err)
	assert.Equal(t, int8(5), savedRule.StatementDay)
	assert.Equal(t, int8(15), savedRule.RepaymentDay)
	assert.Equal(t, "Asia/Shanghai", savedRule.Timezone)

	dueDates, err := Installments.GenerateDueDatesByAccountRule(1710000000, savedRule, 2)
	assert.Nil(t, err)
	assert.Equal(t, []string{"2024-04-15", "2024-05-15"}, dueDates)

	plan := &models.InstallmentPlan{
		Uid:                       installmentTestUid,
		ProviderKey:               models.INSTALLMENT_PROVIDER_HUABEI,
		LiabilityAccountId:        2001,
		DefaultPaymentAccountId:   2002,
		AccountingMode:            models.INSTALLMENT_ACCOUNTING_MODE_REPAYMENT_RECOGNIZED,
		RepaymentCategoryId:       3002,
		TransferCategoryId:        4002,
		Title:                     "Rule Plan",
		PurchaseTime:              1710000000,
		PurchaseTimezoneUtcOffset: 480,
		Currency:                  "CNY",
		PrincipalTotal:            6000,
		FeeTotal:                  0,
		PeriodCount:               2,
		DueDateSource:             models.INSTALLMENT_DUE_DATE_SOURCE_ACCOUNT_RULE,
		StorageMode:               models.INSTALLMENT_STORAGE_MODE_PLAN_ITEMS_ONLY,
		FirstDueDate:              dueDates[0],
		MonthlyDueDay:             15,
	}
	items := []*models.InstallmentItem{
		{SeqNo: 1, DueDate: dueDates[0], PrincipalAmount: 3000, FeeAmount: 0, DueAmount: 3000},
		{SeqNo: 2, DueDate: dueDates[1], PrincipalAmount: 3000, FeeAmount: 0, DueAmount: 3000},
	}

	err = Installments.CreatePlan(ctx, plan, items)
	assert.Nil(t, err)
}

func TestInstallmentService_CreatePlanWithAccountRuleRequiresEnabledRule(t *testing.T) {
	ctx, cleanup := prepareInstallmentServiceTestEnv(t)
	defer cleanup()

	seedInstallmentTestBaseData(t, ctx)

	plan := &models.InstallmentPlan{
		Uid:                       installmentTestUid,
		ProviderKey:               models.INSTALLMENT_PROVIDER_HUABEI,
		LiabilityAccountId:        2001,
		DefaultPaymentAccountId:   2002,
		AccountingMode:            models.INSTALLMENT_ACCOUNTING_MODE_REPAYMENT_RECOGNIZED,
		RepaymentCategoryId:       3002,
		TransferCategoryId:        4002,
		Title:                     "Rule Plan",
		PurchaseTime:              1710000000,
		PurchaseTimezoneUtcOffset: 480,
		Currency:                  "CNY",
		PrincipalTotal:            3000,
		FeeTotal:                  0,
		PeriodCount:               1,
		DueDateSource:             models.INSTALLMENT_DUE_DATE_SOURCE_ACCOUNT_RULE,
		StorageMode:               models.INSTALLMENT_STORAGE_MODE_PLAN_ITEMS_ONLY,
		FirstDueDate:              "2024-04-15",
		MonthlyDueDay:             15,
	}
	items := []*models.InstallmentItem{
		{SeqNo: 1, DueDate: "2024-04-15", PrincipalAmount: 3000, FeeAmount: 0, DueAmount: 3000},
	}

	err := Installments.CreatePlan(ctx, plan, items)
	assert.ErrorIs(t, err, errs.ErrInstallmentAccountRuleNotFound)
}

func TestInstallmentService_GeneratedScheduleTemplatesAreCreatedAndExecuted(t *testing.T) {
	ctx, cleanup := prepareInstallmentServiceTestEnv(t)
	defer cleanup()

	seedInstallmentTestBaseData(t, ctx)

	purchaseTime := time.Date(2024, time.March, 9, 10, 15, 0, 0, time.FixedZone("CST", 8*3600)).Unix()
	plan := &models.InstallmentPlan{
		Uid:                          installmentTestUid,
		ProviderKey:                  models.INSTALLMENT_PROVIDER_HUABEI,
		LiabilityAccountId:           2001,
		DefaultPaymentAccountId:      2002,
		AccountingMode:               models.INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED,
		GeneratedPurchaseTransaction: true,
		PurchaseCategoryId:           3002,
		FeeCategoryId:                3003,
		TransferCategoryId:           4002,
		Title:                        "Phone",
		PurchaseTime:                 purchaseTime,
		PurchaseTimezoneUtcOffset:    480,
		Currency:                     "CNY",
		PrincipalTotal:               9000,
		FeeTotal:                     300,
		PeriodCount:                  3,
		DueDateSource:                models.INSTALLMENT_DUE_DATE_SOURCE_PLAN_RULE,
		StorageMode:                  models.INSTALLMENT_STORAGE_MODE_GENERATED_SCHEDULE_ITEMS,
		FirstDueDate:                 "2024-04-09",
		MonthlyDueDay:                9,
	}
	items := []*models.InstallmentItem{
		{SeqNo: 1, DueDate: "2024-04-09", PrincipalAmount: 3000, FeeAmount: 100, DueAmount: 3100},
		{SeqNo: 2, DueDate: "2024-05-09", PrincipalAmount: 3000, FeeAmount: 100, DueAmount: 3100},
		{SeqNo: 3, DueDate: "2024-06-09", PrincipalAmount: 3000, FeeAmount: 100, DueAmount: 3100},
	}

	err := Installments.CreatePlan(ctx, plan, items)
	assert.Nil(t, err)

	allItems, err := Installments.GetItemsByPlanId(ctx, installmentTestUid, plan.PlanId)
	assert.Nil(t, err)
	assert.NotZero(t, allItems[0].GeneratedTemplateId)
	assert.NotZero(t, allItems[1].GeneratedTemplateId)
	assert.NotZero(t, allItems[2].GeneratedTemplateId)

	template, err := TransactionTemplates.GetTemplateByTemplateId(ctx, installmentTestUid, allItems[0].GeneratedTemplateId)
	assert.Nil(t, err)
	assert.True(t, template.Hidden)
	assert.Equal(t, models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE, template.TemplateType)
	assert.Equal(t, models.TRANSACTION_TYPE_TRANSFER, template.Type)

	err = Transactions.CreateScheduledTransactions(ctx, mustInstallmentDueUnixTime("2024-04-09", 0, 0, 480), time.Minute)
	assert.Nil(t, err)

	updatedItems, err := Installments.GetItemsByPlanId(ctx, installmentTestUid, plan.PlanId)
	assert.Nil(t, err)
	assert.Equal(t, models.INSTALLMENT_ITEM_STATUS_PAID, updatedItems[0].Status)
	assert.NotZero(t, updatedItems[0].RepaymentTransactionId)
	assert.NotZero(t, updatedItems[0].FeeTransactionId)
	assert.Zero(t, updatedItems[0].GeneratedTemplateId)

	_, err = TransactionTemplates.GetTemplateByTemplateId(ctx, installmentTestUid, allItems[0].GeneratedTemplateId)
	assert.ErrorIs(t, err, errs.ErrTransactionTemplateNotFound)
}

func TestInstallmentService_PayGeneratedTemplateItemDeletesTemplate(t *testing.T) {
	ctx, cleanup := prepareInstallmentServiceTestEnv(t)
	defer cleanup()

	seedInstallmentTestBaseData(t, ctx)

	plan := &models.InstallmentPlan{
		Uid:                       installmentTestUid,
		ProviderKey:               models.INSTALLMENT_PROVIDER_JD_BAITIAO,
		LiabilityAccountId:        2001,
		DefaultPaymentAccountId:   2002,
		AccountingMode:            models.INSTALLMENT_ACCOUNTING_MODE_REPAYMENT_RECOGNIZED,
		RepaymentCategoryId:       3002,
		TransferCategoryId:        4002,
		Title:                     "Laptop",
		PurchaseTime:              1710000000,
		PurchaseTimezoneUtcOffset: 480,
		Currency:                  "CNY",
		PrincipalTotal:            6000,
		FeeTotal:                  0,
		PeriodCount:               2,
		DueDateSource:             models.INSTALLMENT_DUE_DATE_SOURCE_PLAN_RULE,
		StorageMode:               models.INSTALLMENT_STORAGE_MODE_GENERATED_SCHEDULE_ITEMS,
		FirstDueDate:              "2024-04-15",
		MonthlyDueDay:             15,
	}
	items := []*models.InstallmentItem{
		{SeqNo: 1, DueDate: "2024-04-15", PrincipalAmount: 3000, FeeAmount: 0, DueAmount: 3000},
		{SeqNo: 2, DueDate: "2024-05-15", PrincipalAmount: 3000, FeeAmount: 0, DueAmount: 3000},
	}

	err := Installments.CreatePlan(ctx, plan, items)
	assert.Nil(t, err)

	allItems, err := Installments.GetItemsByPlanId(ctx, installmentTestUid, plan.PlanId)
	assert.Nil(t, err)
	assert.NotZero(t, allItems[0].GeneratedTemplateId)

	err = Installments.PayItem(ctx, installmentTestUid, &models.InstallmentItemPayRequest{
		ItemId:    allItems[0].ItemId,
		Time:      1710000200,
		UtcOffset: 480,
	})
	assert.Nil(t, err)

	updatedItems, err := Installments.GetItemsByPlanId(ctx, installmentTestUid, plan.PlanId)
	assert.Nil(t, err)
	assert.Equal(t, models.INSTALLMENT_ITEM_STATUS_PAID, updatedItems[0].Status)
	assert.Zero(t, updatedItems[0].GeneratedTemplateId)

	_, err = TransactionTemplates.GetTemplateByTemplateId(ctx, installmentTestUid, allItems[0].GeneratedTemplateId)
	assert.ErrorIs(t, err, errs.ErrTransactionTemplateNotFound)
}

func prepareInstallmentServiceTestEnv(t *testing.T) (core.Context, func()) {
	t.Helper()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "installment-test.db")
	config := &settings.Config{
		DatabaseConfig: &settings.DatabaseConfig{
			DatabaseType: settings.Sqlite3DbType,
			DatabasePath: dbPath,
		},
		UuidGeneratorType: settings.InternalUuidGeneratorType,
		UuidServerId:      1,
	}

	settings.SetCurrentConfig(config)
	assert.Nil(t, datastore.InitializeDataStore(config))
	assert.Nil(t, uuid.InitializeUuidGenerator(config))

	assert.Nil(t, datastore.Container.UserDataStore.SyncStructs(
		new(models.Account),
		new(models.Transaction),
		new(models.TransactionCategory),
		new(models.TransactionTag),
		new(models.TransactionTagIndex),
		new(models.TransactionPictureInfo),
		new(models.TransactionTemplate),
		new(models.InstallmentPlan),
		new(models.InstallmentItem),
		new(models.InstallmentAccountRule),
	))

	ctx := core.NewNullContext()

	return ctx, func() {
		_ = os.Remove(dbPath)
	}
}

func seedInstallmentTestBaseData(t *testing.T, ctx core.Context) {
	t.Helper()

	now := int64(1710000000)
	userDataDB := datastore.Container.UserDataStore.Choose(installmentTestUid)

	accounts := []*models.Account{
		{AccountId: 2001, Uid: installmentTestUid, Category: models.ACCOUNT_CATEGORY_CREDIT_CARD, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, ParentAccountId: 0, Name: "Credit Card", DisplayOrder: 1, Icon: 1, Color: "ff0000", Currency: "CNY", Balance: 0, CreatedUnixTime: now, UpdatedUnixTime: now},
		{AccountId: 2002, Uid: installmentTestUid, Category: models.ACCOUNT_CATEGORY_CHECKING_ACCOUNT, Type: models.ACCOUNT_TYPE_SINGLE_ACCOUNT, ParentAccountId: 0, Name: "Checking", DisplayOrder: 2, Icon: 2, Color: "00ff00", Currency: "CNY", Balance: 100000, CreatedUnixTime: now, UpdatedUnixTime: now},
	}

	categories := []*models.TransactionCategory{
		{CategoryId: 3001, Uid: installmentTestUid, Type: models.CATEGORY_TYPE_EXPENSE, ParentCategoryId: 0, Name: "Expense", DisplayOrder: 1, Icon: 1, Color: "111111", CreatedUnixTime: now, UpdatedUnixTime: now},
		{CategoryId: 3002, Uid: installmentTestUid, Type: models.CATEGORY_TYPE_EXPENSE, ParentCategoryId: 3001, Name: "Shopping", DisplayOrder: 1, Icon: 2, Color: "222222", CreatedUnixTime: now, UpdatedUnixTime: now},
		{CategoryId: 3003, Uid: installmentTestUid, Type: models.CATEGORY_TYPE_EXPENSE, ParentCategoryId: 3001, Name: "Fee", DisplayOrder: 2, Icon: 3, Color: "333333", CreatedUnixTime: now, UpdatedUnixTime: now},
		{CategoryId: 4001, Uid: installmentTestUid, Type: models.CATEGORY_TYPE_TRANSFER, ParentCategoryId: 0, Name: "Transfer", DisplayOrder: 1, Icon: 4, Color: "444444", CreatedUnixTime: now, UpdatedUnixTime: now},
		{CategoryId: 4002, Uid: installmentTestUid, Type: models.CATEGORY_TYPE_TRANSFER, ParentCategoryId: 4001, Name: "Repay", DisplayOrder: 1, Icon: 5, Color: "555555", CreatedUnixTime: now, UpdatedUnixTime: now},
	}

	assert.Nil(t, userDataDB.DoTransaction(ctx, func(sess *xorm.Session) error {
		if _, err := sess.Insert(accounts); err != nil {
			return err
		}

		_, err := sess.Insert(categories)
		return err
	}))
}

func mustInstallmentDueUnixTime(date string, hour int, minute int, utcOffsetMinutes int) int64 {
	location := time.FixedZone("Installment", utcOffsetMinutes*60)
	dueTime, err := time.ParseInLocation("2006-01-02 15:04", date+" "+time.Date(2000, 1, 1, hour, minute, 0, 0, location).Format("15:04"), location)

	if err != nil {
		panic(err)
	}

	return dueTime.Unix()
}
