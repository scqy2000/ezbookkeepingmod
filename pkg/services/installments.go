package services

import (
	"fmt"
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// InstallmentService represents installment service.
type InstallmentService struct {
	ServiceUsingDB
	ServiceUsingUuid
	transactions *TransactionService
	templates    *TransactionTemplateService
}

// Initialize an installment service singleton instance.
var (
	Installments = &InstallmentService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
		transactions: Transactions,
		templates:    TransactionTemplates,
	}
)

// GetAllPlansByUid returns all installment plans.
func (s *InstallmentService) GetAllPlansByUid(c core.Context, uid int64) ([]*models.InstallmentPlan, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var plans []*models.InstallmentPlan
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("purchase_time desc, plan_id desc").Find(&plans)

	return plans, err
}

// GetPlanByPlanId returns one installment plan.
func (s *InstallmentService) GetPlanByPlanId(c core.Context, uid int64, planId int64) (*models.InstallmentPlan, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if planId <= 0 {
		return nil, errs.ErrInstallmentPlanIdInvalid
	}

	plan := &models.InstallmentPlan{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(planId).Where("uid=? AND deleted=?", uid, false).Get(plan)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrInstallmentPlanNotFound
	}

	return plan, nil
}

// GetItemsByPlanId returns all items of one plan.
func (s *InstallmentService) GetItemsByPlanId(c core.Context, uid int64, planId int64) ([]*models.InstallmentItem, error) {
	if _, err := s.GetPlanByPlanId(c, uid, planId); err != nil {
		return nil, err
	}

	var items []*models.InstallmentItem
	err := s.UserDataDB(uid).NewSession(c).Where("plan_id=? AND deleted=?", planId, false).OrderBy("seq_no asc").Find(&items)

	return items, err
}

// GetItemsByPlanIds returns all items grouped by plan id.
func (s *InstallmentService) GetItemsByPlanIds(c core.Context, uid int64, planIds []int64) (map[int64][]*models.InstallmentItem, error) {
	result := make(map[int64][]*models.InstallmentItem)

	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if len(planIds) < 1 {
		return result, nil
	}

	var items []*models.InstallmentItem
	err := s.UserDataDB(uid).NewSession(c).Where("deleted=?", false).In("plan_id", planIds).OrderBy("plan_id asc, seq_no asc").Find(&items)

	if err != nil {
		return nil, err
	}

	for i := 0; i < len(items); i++ {
		result[items[i].PlanId] = append(result[items[i].PlanId], items[i])
	}

	return result, nil
}

// GetAllAccountRulesByUid returns all installment account rules.
func (s *InstallmentService) GetAllAccountRulesByUid(c core.Context, uid int64) ([]*models.InstallmentAccountRule, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var rules []*models.InstallmentAccountRule
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND deleted=?", uid, false).OrderBy("liability_account_id asc, rule_id asc").Find(&rules)

	return rules, err
}

// GetAccountRuleByLiabilityAccountId returns one installment account rule.
func (s *InstallmentService) GetAccountRuleByLiabilityAccountId(c core.Context, uid int64, liabilityAccountId int64) (*models.InstallmentAccountRule, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if liabilityAccountId <= 0 {
		return nil, errs.ErrInstallmentLiabilityAccountInvalid
	}

	return s.getAccountRuleByLiabilityAccountIdInSession(s.UserDataDB(uid).NewSession(c), uid, liabilityAccountId)
}

// SaveAccountRule creates or updates one installment account rule.
func (s *InstallmentService) SaveAccountRule(c core.Context, rule *models.InstallmentAccountRule) error {
	if rule.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if err := rule.Validate(); err != nil {
		return err
	}

	now := time.Now().Unix()
	rule.UpdatedUnixTime = now

	return s.UserDataDB(rule.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		if _, err := s.getValidLiabilityAccount(sess, rule.Uid, rule.LiabilityAccountId); err != nil {
			return err
		}

		existingRule, err := s.getAccountRuleByLiabilityAccountIdInSession(sess, rule.Uid, rule.LiabilityAccountId)

		if err != nil && err != errs.ErrInstallmentAccountRuleNotFound {
			return err
		}

		if existingRule == nil {
			rule.RuleId = s.GenerateUuid(uuid.UUID_TYPE_INSTALLMENT_RULE)

			if rule.RuleId <= 0 {
				return errs.ErrSystemIsBusy
			}

			rule.Deleted = false
			rule.CreatedUnixTime = now
			_, err = sess.Insert(rule)
			return err
		}

		rule.RuleId = existingRule.RuleId
		updatedRows, err := sess.ID(existingRule.RuleId).
			Cols("statement_day", "repayment_day", "timezone", "enabled", "updated_unix_time").
			Where("uid=? AND deleted=?", rule.Uid, false).
			Update(rule)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrInstallmentAccountRuleNotFound
		}

		return nil
	})
}

// DeleteAccountRule deletes one installment account rule.
func (s *InstallmentService) DeleteAccountRule(c core.Context, uid int64, liabilityAccountId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if liabilityAccountId <= 0 {
		return errs.ErrInstallmentLiabilityAccountInvalid
	}

	now := time.Now().Unix()
	updateModel := &models.InstallmentAccountRule{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND liability_account_id=?", uid, false, liabilityAccountId).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrInstallmentAccountRuleNotFound
		}

		return nil
	})
}

// GenerateDueDatesByAccountRule returns due dates initialized by account rule.
func (s *InstallmentService) GenerateDueDatesByAccountRule(purchaseTime int64, rule *models.InstallmentAccountRule, periodCount int16) ([]string, error) {
	if rule == nil {
		return nil, errs.ErrInstallmentAccountRuleNotFound
	}

	if err := rule.Validate(); err != nil {
		return nil, err
	}

	if !rule.Enabled {
		return nil, errs.ErrInstallmentAccountRuleDisabled
	}

	if purchaseTime <= 0 || periodCount <= 0 {
		return nil, errs.ErrInstallmentAccountRuleInvalid
	}

	location, err := time.LoadLocation(rule.Timezone)

	if err != nil {
		return nil, errs.ErrInstallmentTimezoneInvalid
	}

	purchaseLocalTime := time.Unix(purchaseTime, 0).In(location)
	statementDate := s.getFirstStatementDateByRule(purchaseLocalTime, int(rule.StatementDay))
	firstDueDate := s.getFirstRepaymentDateByRule(statementDate, int(rule.StatementDay), int(rule.RepaymentDay), location)
	dueDates := make([]string, periodCount)

	for i := int16(0); i < periodCount; i++ {
		dueDate := s.getLocalDateWithClampedDay(firstDueDate.Year(), firstDueDate.Month()+time.Month(i), int(rule.RepaymentDay), location)
		dueDates[i] = dueDate.Format("2006-01-02")
	}

	return dueDates, nil
}

// CreatePlan saves a new installment plan.
func (s *InstallmentService) CreatePlan(c core.Context, plan *models.InstallmentPlan, items []*models.InstallmentItem) error {
	if plan.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()
	plan.PlanId = s.GenerateUuid(uuid.UUID_TYPE_INSTALLMENT_PLAN)

	if plan.PlanId <= 0 {
		return errs.ErrSystemIsBusy
	}

	plan.Deleted = false
	plan.CreatedUnixTime = now
	plan.UpdatedUnixTime = now

	return s.UserDataDB(plan.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		if err := s.validatePlanForWrite(sess, plan, items); err != nil {
			return err
		}

		if plan.GeneratedPurchaseTransaction {
			purchaseTransaction, err := s.buildPurchaseTransaction(plan)

			if err != nil {
				return err
			}

			if err := s.createTransactionsInSession(c, sess, plan.Uid, []*models.Transaction{purchaseTransaction}); err != nil {
				return err
			}

			plan.PurchaseTransactionId = purchaseTransaction.TransactionId
		}

		if _, err := sess.Insert(plan); err != nil {
			return err
		}

		insertItems, err := s.insertPlanItems(sess, plan.PlanId, items)

		if err != nil {
			return err
		}

		return s.createGeneratedTemplatesInSession(sess, plan, insertItems)
	})
}

// ModifyPlan updates an installment plan.
func (s *InstallmentService) ModifyPlan(c core.Context, plan *models.InstallmentPlan, items []*models.InstallmentItem) error {
	if plan.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	now := time.Now().Unix()
	plan.UpdatedUnixTime = now

	return s.UserDataDB(plan.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		oldPlan, err := s.getPlanByPlanIdInSession(sess, plan.Uid, plan.PlanId)

		if err != nil {
			return err
		}

		hasPaidItems, err := s.hasPaidItems(sess, plan.PlanId)

		if err != nil {
			return err
		} else if hasPaidItems {
			return errs.ErrInstallmentHasPaidItemsCannotModify
		}

		oldItems, err := s.getItemsByPlanIdInSession(sess, plan.PlanId)

		if err != nil {
			return err
		}

		if err := s.validatePlanForWrite(sess, plan, items); err != nil {
			return err
		}

		if oldPlan.GeneratedPurchaseTransaction && oldPlan.PurchaseTransactionId > 0 {
			if err := s.deleteTransactionInSession(c, sess, plan.Uid, oldPlan.PurchaseTransactionId); err != nil {
				return err
			}
		}

		if err := s.deleteGeneratedTemplatesInSession(sess, plan.Uid, oldItems); err != nil {
			return err
		}

		if plan.GeneratedPurchaseTransaction {
			purchaseTransaction, err := s.buildPurchaseTransaction(plan)

			if err != nil {
				return err
			}

			if err := s.createTransactionsInSession(c, sess, plan.Uid, []*models.Transaction{purchaseTransaction}); err != nil {
				return err
			}

			plan.PurchaseTransactionId = purchaseTransaction.TransactionId
		}

		updatedRows, err := sess.ID(plan.PlanId).
			Cols("provider_key", "custom_provider_name", "liability_account_id", "default_payment_account_id", "accounting_mode",
				"purchase_transaction_id", "generated_purchase_transaction", "purchase_category_id", "repayment_category_id",
				"fee_category_id", "transfer_category_id", "title", "notes", "purchase_time", "purchase_timezone_utc_offset",
				"currency", "principal_total", "fee_total", "period_count", "due_date_source", "storage_mode", "first_due_date",
				"monthly_due_day", "updated_unix_time").
			Where("uid=? AND deleted=?", plan.Uid, false).
			Update(plan)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrInstallmentPlanNotFound
		}

		itemDeleteModel := &models.InstallmentItem{
			Deleted:         true,
			DeletedUnixTime: now,
		}

		if _, err := sess.Cols("deleted", "deleted_unix_time").Where("plan_id=? AND deleted=?", plan.PlanId, false).Update(itemDeleteModel); err != nil {
			return err
		}

		insertItems, err := s.insertPlanItems(sess, plan.PlanId, items)

		if err != nil {
			return err
		}

		return s.createGeneratedTemplatesInSession(sess, plan, insertItems)
	})
}

// DeletePlan deletes an installment plan.
func (s *InstallmentService) DeletePlan(c core.Context, uid int64, planId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if planId <= 0 {
		return errs.ErrInstallmentPlanIdInvalid
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		plan, err := s.getPlanByPlanIdInSession(sess, uid, planId)

		if err != nil {
			return err
		}

		hasPaidItems, err := s.hasPaidItems(sess, planId)

		if err != nil {
			return err
		} else if hasPaidItems {
			return errs.ErrInstallmentHasPaidItemsCannotDelete
		}

		items, err := s.getItemsByPlanIdInSession(sess, planId)

		if err != nil {
			return err
		}

		planDeleteModel := &models.InstallmentPlan{
			Deleted:         true,
			DeletedUnixTime: now,
		}

		itemDeleteModel := &models.InstallmentItem{
			Deleted:         true,
			DeletedUnixTime: now,
		}

		if _, err := sess.ID(planId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(planDeleteModel); err != nil {
			return err
		}

		if _, err := sess.Cols("deleted", "deleted_unix_time").Where("plan_id=? AND deleted=?", planId, false).Update(itemDeleteModel); err != nil {
			return err
		}

		if err := s.deleteGeneratedTemplatesInSession(sess, uid, items); err != nil {
			return err
		}

		if plan.GeneratedPurchaseTransaction && plan.PurchaseTransactionId > 0 {
			return s.deleteTransactionInSession(c, sess, uid, plan.PurchaseTransactionId)
		}

		return nil
	})
}

// PayItem marks an installment item as paid and creates related transactions.
func (s *InstallmentService) PayItem(c core.Context, uid int64, req *models.InstallmentItemPayRequest) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if req.ItemId <= 0 {
		return errs.ErrInstallmentItemIdInvalid
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		plan, item, err := s.getPlanAndItemByItemIdInSession(sess, uid, req.ItemId)

		if err != nil {
			return err
		}

		if item.Status == models.INSTALLMENT_ITEM_STATUS_PAID {
			return errs.ErrInstallmentItemAlreadyPaid
		}

		paymentAccountId := req.PaymentAccountId

		if paymentAccountId <= 0 {
			paymentAccountId = plan.DefaultPaymentAccountId
		}

		if paymentAccountId <= 0 {
			return errs.ErrInstallmentPaymentAccountRequired
		}

		paymentAccount, err := s.getValidPaymentAccount(sess, uid, paymentAccountId, plan.Currency)

		if err != nil {
			return err
		}

		comment := req.Comment

		if comment == "" {
			comment = fmt.Sprintf("%s #%d", plan.Title, item.SeqNo)
		}

		transactions, expenseTransaction, repaymentTransaction, feeTransaction, err := s.buildRepaymentTransactions(plan, item, paymentAccount, req, comment)

		if err != nil {
			return err
		}

		if err := s.createTransactionsInSession(c, sess, uid, transactions); err != nil {
			return err
		}

		if item.GeneratedTemplateId > 0 {
			if err := s.deleteGeneratedTemplateByIdInSession(sess, uid, item.GeneratedTemplateId); err != nil {
				return err
			}

			item.GeneratedTemplateId = 0
		}

		item.Status = models.INSTALLMENT_ITEM_STATUS_PAID
		item.PaidTime = req.Time
		item.UpdatedUnixTime = now

		if expenseTransaction != nil {
			item.ExpenseTransactionId = expenseTransaction.TransactionId
		}

		if repaymentTransaction != nil {
			item.RepaymentTransactionId = repaymentTransaction.TransactionId
		}

		if feeTransaction != nil {
			item.FeeTransactionId = feeTransaction.TransactionId
		}

		updatedRows, err := sess.ID(item.ItemId).Cols("status", "paid_time", "expense_transaction_id", "repayment_transaction_id", "fee_transaction_id", "generated_template_id", "updated_unix_time").Update(item)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrInstallmentItemNotFound
		}

		return nil
	})
}

// UnpayItem reverts an installment payment.
func (s *InstallmentService) UnpayItem(c core.Context, uid int64, itemId int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if itemId <= 0 {
		return errs.ErrInstallmentItemIdInvalid
	}

	now := time.Now().Unix()

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		plan, item, err := s.getPlanAndItemByItemIdInSession(sess, uid, itemId)

		if err != nil {
			return err
		}

		if item.Status != models.INSTALLMENT_ITEM_STATUS_PAID {
			return errs.ErrInstallmentItemNotPaid
		}

		for _, transactionId := range []int64{item.FeeTransactionId, item.ExpenseTransactionId, item.RepaymentTransactionId} {
			if transactionId > 0 {
				if err := s.deleteTransactionInSession(c, sess, uid, transactionId); err != nil {
					return err
				}
			}
		}

		item.Status = models.INSTALLMENT_ITEM_STATUS_UNPAID
		item.PaidTime = 0
		item.ExpenseTransactionId = 0
		item.RepaymentTransactionId = 0
		item.FeeTransactionId = 0
		item.UpdatedUnixTime = now

		updatedRows, err := sess.ID(item.ItemId).Cols("status", "paid_time", "expense_transaction_id", "repayment_transaction_id", "fee_transaction_id", "updated_unix_time").Update(item)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrInstallmentItemNotFound
		}

		return s.createGeneratedTemplatesInSession(sess, plan, []*models.InstallmentItem{item})
	})
}

// HandleScheduledTemplateExecution updates installment state after a generated template creates a transaction.
func (s *InstallmentService) HandleScheduledTemplateExecution(c core.Context, sess *xorm.Session, template *models.TransactionTemplate, repaymentTransaction *models.Transaction) error {
	if template == nil || repaymentTransaction == nil || template.TemplateId <= 0 {
		return nil
	}

	item, err := s.getItemByGeneratedTemplateIdInSession(sess, template.Uid, template.TemplateId)

	if err == errs.ErrInstallmentItemNotFound {
		return nil
	} else if err != nil {
		return err
	}

	plan, err := s.getPlanByPlanIdInSession(sess, template.Uid, item.PlanId)

	if err != nil {
		return err
	}

	if item.Status == models.INSTALLMENT_ITEM_STATUS_PAID {
		if err := s.deleteGeneratedTemplateByIdInSession(sess, template.Uid, template.TemplateId); err != nil {
			return err
		}

		return nil
	}

	extraTransactions, expenseTransaction, feeTransaction, err := s.buildSupplementalRepaymentTransactions(plan, item, repaymentTransaction)

	if err != nil {
		return err
	}

	if err := s.createTransactionsInSession(c, sess, template.Uid, extraTransactions); err != nil {
		return err
	}

	now := time.Now().Unix()
	item.Status = models.INSTALLMENT_ITEM_STATUS_PAID
	item.PaidTime = utils.GetUnixTimeFromTransactionTime(repaymentTransaction.TransactionTime)
	item.RepaymentTransactionId = repaymentTransaction.TransactionId
	item.GeneratedTemplateId = 0
	item.UpdatedUnixTime = now

	if expenseTransaction != nil {
		item.ExpenseTransactionId = expenseTransaction.TransactionId
	}

	if feeTransaction != nil {
		item.FeeTransactionId = feeTransaction.TransactionId
	}

	if err := s.deleteGeneratedTemplateByIdInSession(sess, template.Uid, template.TemplateId); err != nil {
		return err
	}

	updatedRows, err := sess.ID(item.ItemId).Cols("status", "paid_time", "expense_transaction_id", "repayment_transaction_id", "fee_transaction_id", "generated_template_id", "updated_unix_time").Update(item)

	if err != nil {
		return err
	} else if updatedRows < 1 {
		return errs.ErrInstallmentItemNotFound
	}

	return nil
}

func (s *InstallmentService) getPlanByPlanIdInSession(sess *xorm.Session, uid int64, planId int64) (*models.InstallmentPlan, error) {
	plan := &models.InstallmentPlan{}
	has, err := sess.ID(planId).Where("uid=? AND deleted=?", uid, false).Get(plan)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrInstallmentPlanNotFound
	}

	return plan, nil
}

func (s *InstallmentService) getPlanAndItemByItemIdInSession(sess *xorm.Session, uid int64, itemId int64) (*models.InstallmentPlan, *models.InstallmentItem, error) {
	item := &models.InstallmentItem{}
	has, err := sess.ID(itemId).Where("deleted=?", false).Get(item)

	if err != nil {
		return nil, nil, err
	} else if !has {
		return nil, nil, errs.ErrInstallmentItemNotFound
	}

	plan, err := s.getPlanByPlanIdInSession(sess, uid, item.PlanId)

	if err != nil {
		return nil, nil, err
	}

	return plan, item, nil
}

func (s *InstallmentService) hasPaidItems(sess *xorm.Session, planId int64) (bool, error) {
	return sess.Where("plan_id=? AND deleted=? AND status=?", planId, false, models.INSTALLMENT_ITEM_STATUS_PAID).Limit(1).Exist(&models.InstallmentItem{})
}

func (s *InstallmentService) validatePlanForWrite(sess *xorm.Session, plan *models.InstallmentPlan, items []*models.InstallmentItem) error {
	if err := plan.Validate(items); err != nil {
		return err
	}

	liabilityAccount, err := s.getValidLiabilityAccount(sess, plan.Uid, plan.LiabilityAccountId)

	if err != nil {
		return err
	}

	if liabilityAccount.Currency != plan.Currency {
		return errs.ErrInstallmentCurrencyInvalid
	}

	if plan.DefaultPaymentAccountId > 0 {
		if _, err := s.getValidPaymentAccount(sess, plan.Uid, plan.DefaultPaymentAccountId, plan.Currency); err != nil {
			return err
		}
	} else if plan.StorageMode == models.INSTALLMENT_STORAGE_MODE_GENERATED_SCHEDULE_ITEMS {
		return errs.ErrInstallmentPaymentAccountRequired
	}

	if plan.DueDateSource == models.INSTALLMENT_DUE_DATE_SOURCE_ACCOUNT_RULE {
		accountRule, err := s.getAccountRuleByLiabilityAccountIdInSession(sess, plan.Uid, plan.LiabilityAccountId)

		if err != nil {
			return err
		}

		if !accountRule.Enabled {
			return errs.ErrInstallmentAccountRuleDisabled
		}
	}

	if err := s.validateCategory(sess, plan.Uid, plan.TransferCategoryId, models.CATEGORY_TYPE_TRANSFER, errs.ErrInstallmentTransferCategoryRequired); err != nil {
		return err
	}

	if plan.AccountingMode == models.INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED {
		if plan.PurchaseTransactionId > 0 && plan.GeneratedPurchaseTransaction {
			return errs.ErrInstallmentPurchaseTransactionInvalid
		}

		if plan.PurchaseTransactionId <= 0 {
			if !plan.GeneratedPurchaseTransaction {
				return errs.ErrInstallmentPurchaseTransactionRequired
			}

			if err := s.validateCategory(sess, plan.Uid, plan.PurchaseCategoryId, models.CATEGORY_TYPE_EXPENSE, errs.ErrInstallmentPurchaseCategoryRequired); err != nil {
				return err
			}
		} else {
			if err := s.validatePurchaseTransaction(sess, plan, liabilityAccount); err != nil {
				return err
			}
		}

		if plan.FeeTotal > 0 {
			if err := s.validateCategory(sess, plan.Uid, plan.FeeCategoryId, models.CATEGORY_TYPE_EXPENSE, errs.ErrInstallmentFeeCategoryRequired); err != nil {
				return err
			}
		}
	} else if plan.AccountingMode == models.INSTALLMENT_ACCOUNTING_MODE_REPAYMENT_RECOGNIZED {
		if plan.GeneratedPurchaseTransaction {
			return errs.ErrInstallmentOnlyPurchaseRecognizedCanGenerate
		}

		if plan.PurchaseTransactionId > 0 {
			return errs.ErrInstallmentPurchaseTransactionInvalid
		}

		if err := s.validateCategory(sess, plan.Uid, plan.RepaymentCategoryId, models.CATEGORY_TYPE_EXPENSE, errs.ErrInstallmentRepaymentCategoryRequired); err != nil {
			return err
		}
	} else {
		return errs.ErrInstallmentAccountingModeInvalid
	}

	return nil
}

func (s *InstallmentService) getAccountRuleByLiabilityAccountIdInSession(sess *xorm.Session, uid int64, liabilityAccountId int64) (*models.InstallmentAccountRule, error) {
	rule := &models.InstallmentAccountRule{}
	has, err := sess.Where("uid=? AND deleted=? AND liability_account_id=?", uid, false, liabilityAccountId).Limit(1).Get(rule)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrInstallmentAccountRuleNotFound
	}

	return rule, nil
}

func (s *InstallmentService) getItemsByPlanIdInSession(sess *xorm.Session, planId int64) ([]*models.InstallmentItem, error) {
	var items []*models.InstallmentItem
	err := sess.Where("plan_id=? AND deleted=?", planId, false).OrderBy("seq_no asc").Find(&items)
	return items, err
}

func (s *InstallmentService) getItemByGeneratedTemplateIdInSession(sess *xorm.Session, uid int64, templateId int64) (*models.InstallmentItem, error) {
	item := &models.InstallmentItem{}
	has, err := sess.Where("deleted=? AND generated_template_id>0 AND generated_template_id=?", false, templateId).Limit(1).Get(item)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrInstallmentItemNotFound
	}

	plan, err := s.getPlanByPlanIdInSession(sess, uid, item.PlanId)

	if err != nil {
		return nil, err
	}

	if plan.Uid != uid {
		return nil, errs.ErrInstallmentItemNotFound
	}

	return item, nil
}

func (s *InstallmentService) validatePurchaseTransaction(sess *xorm.Session, plan *models.InstallmentPlan, liabilityAccount *models.Account) error {
	transaction := &models.Transaction{}
	has, err := sess.ID(plan.PurchaseTransactionId).Where("uid=? AND deleted=?", plan.Uid, false).Get(transaction)

	if err != nil {
		return err
	} else if !has {
		return errs.ErrInstallmentPurchaseTransactionInvalid
	}

	if transaction.Type != models.TRANSACTION_DB_TYPE_EXPENSE ||
		transaction.AccountId != liabilityAccount.AccountId ||
		transaction.Amount != plan.PrincipalTotal {
		return errs.ErrInstallmentPurchaseTransactionInvalid
	}

	return nil
}

func (s *InstallmentService) validateCategory(sess *xorm.Session, uid int64, categoryId int64, categoryType models.TransactionCategoryType, requiredErr *errs.Error) error {
	if categoryId <= 0 {
		return requiredErr
	}

	category := &models.TransactionCategory{}
	has, err := sess.ID(categoryId).Where("uid=? AND deleted=?", uid, false).Get(category)

	if err != nil {
		return err
	} else if !has || category.Hidden || category.ParentCategoryId == models.LevelOneTransactionCategoryParentId || category.Type != categoryType {
		return errs.ErrTransactionCategoryTypeInvalid
	}

	return nil
}

func (s *InstallmentService) getValidLiabilityAccount(sess *xorm.Session, uid int64, accountId int64) (*models.Account, error) {
	account := &models.Account{}
	has, err := sess.ID(accountId).Where("uid=? AND deleted=?", uid, false).Get(account)

	if err != nil {
		return nil, err
	} else if !has || account.Hidden || account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || !account.Category.IsLiability() {
		return nil, errs.ErrInstallmentLiabilityAccountInvalid
	}

	return account, nil
}

func (s *InstallmentService) getValidPaymentAccount(sess *xorm.Session, uid int64, accountId int64, expectedCurrency string) (*models.Account, error) {
	account := &models.Account{}
	has, err := sess.ID(accountId).Where("uid=? AND deleted=?", uid, false).Get(account)

	if err != nil {
		return nil, err
	} else if !has || account.Hidden || account.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || !account.Category.IsAsset() {
		return nil, errs.ErrInstallmentPaymentAccountInvalid
	}

	if account.Currency != expectedCurrency {
		return nil, errs.ErrInstallmentCurrencyInvalid
	}

	return account, nil
}

func (s *InstallmentService) buildPurchaseTransaction(plan *models.InstallmentPlan) (*models.Transaction, error) {
	if plan.AccountingMode != models.INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED {
		return nil, errs.ErrInstallmentOnlyPurchaseRecognizedCanGenerate
	}

	return &models.Transaction{
		Uid:               plan.Uid,
		Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
		CategoryId:        plan.PurchaseCategoryId,
		AccountId:         plan.LiabilityAccountId,
		TransactionTime:   utils.GetMinTransactionTimeFromUnixTime(plan.PurchaseTime),
		TimezoneUtcOffset: plan.PurchaseTimezoneUtcOffset,
		Amount:            plan.PrincipalTotal,
		Comment:           plan.Title,
	}, nil
}

func (s *InstallmentService) buildRepaymentTransactions(plan *models.InstallmentPlan, item *models.InstallmentItem, paymentAccount *models.Account, req *models.InstallmentItemPayRequest, comment string) ([]*models.Transaction, *models.Transaction, *models.Transaction, *models.Transaction, error) {
	baseTransactionTime := utils.GetMinTransactionTimeFromUnixTime(req.Time)
	transactions := make([]*models.Transaction, 0, 2)
	var expenseTransaction *models.Transaction
	var repaymentTransaction *models.Transaction
	var feeTransaction *models.Transaction

	if plan.AccountingMode == models.INSTALLMENT_ACCOUNTING_MODE_REPAYMENT_RECOGNIZED {
		repaymentCategoryId := plan.RepaymentCategoryId

		if req.RepaymentCategoryId > 0 {
			repaymentCategoryId = req.RepaymentCategoryId
		}

		expenseTransaction = &models.Transaction{
			Uid:               plan.Uid,
			Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
			CategoryId:        repaymentCategoryId,
			AccountId:         plan.LiabilityAccountId,
			TransactionTime:   baseTransactionTime,
			TimezoneUtcOffset: req.UtcOffset,
			Amount:            item.DueAmount,
			Comment:           comment,
		}
		transactions = append(transactions, expenseTransaction)
	}

	if plan.AccountingMode == models.INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED && item.FeeAmount > 0 {
		feeCategoryId := plan.FeeCategoryId

		if req.FeeCategoryId > 0 {
			feeCategoryId = req.FeeCategoryId
		}

		feeTransaction = &models.Transaction{
			Uid:               plan.Uid,
			Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
			CategoryId:        feeCategoryId,
			AccountId:         plan.LiabilityAccountId,
			TransactionTime:   baseTransactionTime,
			TimezoneUtcOffset: req.UtcOffset,
			Amount:            item.FeeAmount,
			Comment:           comment,
		}
		transactions = append(transactions, feeTransaction)
	}

	transferCategoryId := plan.TransferCategoryId

	if req.TransferCategoryId > 0 {
		transferCategoryId = req.TransferCategoryId
	}

	repaymentTransaction = &models.Transaction{
		Uid:                  plan.Uid,
		Type:                 models.TRANSACTION_DB_TYPE_TRANSFER_OUT,
		CategoryId:           transferCategoryId,
		AccountId:            paymentAccount.AccountId,
		TransactionTime:      baseTransactionTime + int64(len(transactions))*2,
		TimezoneUtcOffset:    req.UtcOffset,
		Amount:               item.DueAmount,
		RelatedAccountId:     plan.LiabilityAccountId,
		RelatedAccountAmount: item.DueAmount,
		Comment:              comment,
	}
	transactions = append(transactions, repaymentTransaction)

	return transactions, expenseTransaction, repaymentTransaction, feeTransaction, nil
}

func (s *InstallmentService) buildSupplementalRepaymentTransactions(plan *models.InstallmentPlan, item *models.InstallmentItem, repaymentTransaction *models.Transaction) ([]*models.Transaction, *models.Transaction, *models.Transaction, error) {
	transactions := make([]*models.Transaction, 0, 2)
	comment := repaymentTransaction.Comment
	var expenseTransaction *models.Transaction
	var feeTransaction *models.Transaction

	if comment == "" {
		comment = fmt.Sprintf("%s #%d", plan.Title, item.SeqNo)
	}

	if plan.AccountingMode == models.INSTALLMENT_ACCOUNTING_MODE_REPAYMENT_RECOGNIZED {
		expenseTransaction = &models.Transaction{
			Uid:               plan.Uid,
			Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
			CategoryId:        plan.RepaymentCategoryId,
			AccountId:         plan.LiabilityAccountId,
			TransactionTime:   repaymentTransaction.TransactionTime,
			TimezoneUtcOffset: repaymentTransaction.TimezoneUtcOffset,
			Amount:            item.DueAmount,
			Comment:           comment,
		}
		transactions = append(transactions, expenseTransaction)
	}

	if plan.AccountingMode == models.INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED && item.FeeAmount > 0 {
		feeTransaction = &models.Transaction{
			Uid:               plan.Uid,
			Type:              models.TRANSACTION_DB_TYPE_EXPENSE,
			CategoryId:        plan.FeeCategoryId,
			AccountId:         plan.LiabilityAccountId,
			TransactionTime:   repaymentTransaction.TransactionTime,
			TimezoneUtcOffset: repaymentTransaction.TimezoneUtcOffset,
			Amount:            item.FeeAmount,
			Comment:           comment,
		}
		transactions = append(transactions, feeTransaction)
	}

	return transactions, expenseTransaction, feeTransaction, nil
}

func (s *InstallmentService) insertPlanItems(sess *xorm.Session, planId int64, items []*models.InstallmentItem) ([]*models.InstallmentItem, error) {
	now := time.Now().Unix()
	itemIds := s.GenerateUuids(uuid.UUID_TYPE_INSTALLMENT_ITEM, uint16(len(items)))

	if len(itemIds) < len(items) {
		return nil, errs.ErrSystemIsBusy
	}

	insertItems := make([]*models.InstallmentItem, len(items))

	for i := 0; i < len(items); i++ {
		insertItems[i] = &models.InstallmentItem{
			ItemId:          itemIds[i],
			PlanId:          planId,
			Deleted:         false,
			SeqNo:           items[i].SeqNo,
			DueDate:         items[i].DueDate,
			PrincipalAmount: items[i].PrincipalAmount,
			FeeAmount:       items[i].FeeAmount,
			DueAmount:       items[i].DueAmount,
			Status:          models.INSTALLMENT_ITEM_STATUS_UNPAID,
			CreatedUnixTime: now,
			UpdatedUnixTime: now,
		}
	}

	_, err := sess.Insert(insertItems)
	return insertItems, err
}

func (s *InstallmentService) createGeneratedTemplatesInSession(sess *xorm.Session, plan *models.InstallmentPlan, items []*models.InstallmentItem) error {
	if plan == nil || plan.StorageMode != models.INSTALLMENT_STORAGE_MODE_GENERATED_SCHEDULE_ITEMS || len(items) < 1 {
		return nil
	}

	if plan.DefaultPaymentAccountId <= 0 {
		return errs.ErrInstallmentPaymentAccountRequired
	}

	now := time.Now().Unix()
	templateIds := s.templates.GenerateUuids(uuid.UUID_TYPE_TEMPLATE, uint16(len(items)))
	scheduledTimezoneUtcOffset := plan.PurchaseTimezoneUtcOffset

	if plan.DueDateSource == models.INSTALLMENT_DUE_DATE_SOURCE_ACCOUNT_RULE {
		accountRule, err := s.getAccountRuleByLiabilityAccountIdInSession(sess, plan.Uid, plan.LiabilityAccountId)

		if err != nil {
			return err
		}

		scheduledTimezoneUtcOffset, err = s.getTimezoneUtcOffsetByRuleDate(items[0].DueDate, accountRule.Timezone)

		if err != nil {
			return err
		}
	}

	if len(templateIds) < len(items) {
		return errs.ErrSystemIsBusy
	}

	insertTemplates := make([]*models.TransactionTemplate, 0, len(items))

	for i := 0; i < len(items); i++ {
		item := items[i]

		if item.Status == models.INSTALLMENT_ITEM_STATUS_PAID {
			continue
		}

		template, err := s.buildGeneratedTemplate(plan, item, templateIds[i], now, scheduledTimezoneUtcOffset)

		if err != nil {
			return err
		}

		if err := s.templates.isTemplateValid(sess, template); err != nil {
			return err
		}

		insertTemplates = append(insertTemplates, template)
		item.GeneratedTemplateId = template.TemplateId
		item.UpdatedUnixTime = now
	}

	if len(insertTemplates) < 1 {
		return nil
	}

	if _, err := sess.Insert(insertTemplates); err != nil {
		return err
	}

	for i := 0; i < len(items); i++ {
		if items[i].GeneratedTemplateId <= 0 {
			continue
		}

		if _, err := sess.ID(items[i].ItemId).Cols("generated_template_id", "updated_unix_time").Update(items[i]); err != nil {
			return err
		}
	}

	return nil
}

func (s *InstallmentService) buildGeneratedTemplate(plan *models.InstallmentPlan, item *models.InstallmentItem, templateId int64, now int64, scheduledTimezoneUtcOffset int16) (*models.TransactionTemplate, error) {
	startTime, err := utils.ParseFromLongDateFirstTime(item.DueDate, scheduledTimezoneUtcOffset)

	if err != nil {
		return nil, errs.ErrInstallmentDateInvalid
	}

	endTime, err := utils.ParseFromLongDateLastTime(item.DueDate, scheduledTimezoneUtcOffset)

	if err != nil {
		return nil, errs.ErrInstallmentDateInvalid
	}

	monthlyDay, err := s.getMonthDayFromDate(item.DueDate)

	if err != nil {
		return nil, err
	}

	comment := fmt.Sprintf("%s #%d", plan.Title, item.SeqNo)
	startUnixTime := startTime.Unix()
	endUnixTime := endTime.Unix()
	return &models.TransactionTemplate{
		TemplateId:                 templateId,
		Uid:                        plan.Uid,
		Deleted:                    false,
		TemplateType:               models.TRANSACTION_TEMPLATE_TYPE_SCHEDULE,
		Name:                       comment,
		Type:                       models.TRANSACTION_TYPE_TRANSFER,
		CategoryId:                 plan.TransferCategoryId,
		AccountId:                  plan.DefaultPaymentAccountId,
		ScheduledFrequencyType:     models.TRANSACTION_SCHEDULE_FREQUENCY_TYPE_MONTHLY,
		ScheduledFrequency:         monthlyDay,
		ScheduledStartTime:         &startUnixTime,
		ScheduledEndTime:           &endUnixTime,
		ScheduledAt:                s.getUTCScheduledAt(scheduledTimezoneUtcOffset),
		ScheduledTimezoneUtcOffset: scheduledTimezoneUtcOffset,
		TagIds:                     "",
		Amount:                     item.DueAmount,
		RelatedAccountId:           plan.LiabilityAccountId,
		RelatedAccountAmount:       item.DueAmount,
		HideAmount:                 false,
		Comment:                    comment,
		DisplayOrder:               0,
		Hidden:                     true,
		CreatedUnixTime:            now,
		UpdatedUnixTime:            now,
	}, nil
}

func (s *InstallmentService) deleteGeneratedTemplatesInSession(sess *xorm.Session, uid int64, items []*models.InstallmentItem) error {
	templateIds := make([]int64, 0, len(items))

	for i := 0; i < len(items); i++ {
		if items[i].GeneratedTemplateId > 0 {
			templateIds = append(templateIds, items[i].GeneratedTemplateId)
		}
	}

	return s.deleteGeneratedTemplateIdsInSession(sess, uid, templateIds)
}

func (s *InstallmentService) deleteGeneratedTemplateByIdInSession(sess *xorm.Session, uid int64, templateId int64) error {
	if templateId <= 0 {
		return nil
	}

	return s.deleteGeneratedTemplateIdsInSession(sess, uid, []int64{templateId})
}

func (s *InstallmentService) deleteGeneratedTemplateIdsInSession(sess *xorm.Session, uid int64, templateIds []int64) error {
	if len(templateIds) < 1 {
		return nil
	}

	now := time.Now().Unix()
	updateModel := &models.TransactionTemplate{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	_, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).In("template_id", templateIds).Update(updateModel)
	return err
}

func (s *InstallmentService) getUTCScheduledAt(scheduledTimezoneUtcOffset int16) int16 {
	templateTimeZone := time.FixedZone("Template Timezone", int(scheduledTimezoneUtcOffset)*60)
	transactionTime := time.Date(2020, 1, 1, 0, 0, 0, 0, templateTimeZone)
	transactionTimeInUTC := transactionTime.In(time.UTC)
	return int16(transactionTimeInUTC.Hour()*60 + transactionTimeInUTC.Minute())
}

func (s *InstallmentService) getTimezoneUtcOffsetByRuleDate(date string, timezoneName string) (int16, error) {
	location, err := time.LoadLocation(timezoneName)

	if err != nil {
		return 0, errs.ErrInstallmentTimezoneInvalid
	}

	localTime, err := time.ParseInLocation("2006-01-02", date, location)

	if err != nil {
		return 0, errs.ErrInstallmentDateInvalid
	}

	_, secondsOffset := localTime.Zone()
	return int16(secondsOffset / 60), nil
}

func (s *InstallmentService) getMonthDayFromDate(date string) (string, error) {
	parsedDate, err := time.Parse("2006-01-02", date)

	if err != nil {
		return "", errs.ErrInstallmentDateInvalid
	}

	return utils.Int64ToString(int64(parsedDate.Day())), nil
}

func (s *InstallmentService) getFirstStatementDateByRule(purchaseLocalTime time.Time, statementDay int) time.Time {
	statementDate := s.getLocalDateWithClampedDay(purchaseLocalTime.Year(), purchaseLocalTime.Month(), statementDay, purchaseLocalTime.Location())

	if purchaseLocalTime.Day() > statementDate.Day() {
		statementDate = s.getLocalDateWithClampedDay(purchaseLocalTime.Year(), purchaseLocalTime.Month()+1, statementDay, purchaseLocalTime.Location())
	}

	return statementDate
}

func (s *InstallmentService) getFirstRepaymentDateByRule(statementDate time.Time, statementDay int, repaymentDay int, location *time.Location) time.Time {
	if repaymentDay > statementDay {
		return s.getLocalDateWithClampedDay(statementDate.Year(), statementDate.Month(), repaymentDay, location)
	}

	return s.getLocalDateWithClampedDay(statementDate.Year(), statementDate.Month()+1, repaymentDay, location)
}

func (s *InstallmentService) getLocalDateWithClampedDay(year int, month time.Month, day int, location *time.Location) time.Time {
	firstDayOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, location)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1).Day()

	if day > lastDayOfMonth {
		day = lastDayOfMonth
	}

	return time.Date(firstDayOfMonth.Year(), firstDayOfMonth.Month(), day, 0, 0, 0, 0, location)
}

func (s *InstallmentService) createTransactionsInSession(c core.Context, sess *xorm.Session, uid int64, transactions []*models.Transaction) error {
	now := time.Now().Unix()
	needTransactionUuidCount := uint16(0)

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if transaction.Uid != uid {
			return errs.ErrUserIdInvalid
		}

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			needTransactionUuidCount += 2
		} else {
			needTransactionUuidCount++
		}

		transaction.CreatedUnixTime = now
		transaction.UpdatedUnixTime = now
	}

	transactionUuids := s.transactions.GenerateUuids(uuid.UUID_TYPE_TRANSACTION, needTransactionUuidCount)
	transactionUuidIndex := 0

	if len(transactionUuids) < int(needTransactionUuidCount) {
		return errs.ErrSystemIsBusy
	}

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]
		transaction.TransactionId = transactionUuids[transactionUuidIndex]
		transactionUuidIndex++

		if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
			transaction.RelatedId = transactionUuids[transactionUuidIndex]
			transactionUuidIndex++
		}
	}

	userDataDb := s.UserDataDB(uid)

	for i := 0; i < len(transactions); i++ {
		if err := s.transactions.doCreateTransaction(c, userDataDb, sess, transactions[i], nil, nil, nil, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *InstallmentService) deleteTransactionInSession(c core.Context, sess *xorm.Session, uid int64, transactionId int64) error {
	now := time.Now().Unix()

	updateModel := &models.Transaction{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	tagIndexUpdateModel := &models.TransactionTagIndex{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	pictureUpdateModel := &models.TransactionPictureInfo{
		Deleted:         true,
		DeletedUnixTime: now,
	}

	oldTransaction := &models.Transaction{}
	has, err := sess.ID(transactionId).Where("uid=? AND deleted=?", uid, false).Get(oldTransaction)

	if err != nil {
		return err
	} else if !has {
		return errs.ErrTransactionNotFound
	}

	sourceAccount, destinationAccount, err := s.transactions.getAccountModels(sess, oldTransaction)

	if err != nil {
		return err
	}

	if sourceAccount.Hidden || (destinationAccount != nil && destinationAccount.Hidden) {
		return errs.ErrCannotDeleteTransactionInHiddenAccount
	}

	if sourceAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS || (destinationAccount != nil && destinationAccount.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS) {
		return errs.ErrCannotDeleteTransactionInParentAccount
	}

	deletedRows, err := sess.ID(oldTransaction.TransactionId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

	if err != nil {
		return err
	} else if deletedRows < 1 {
		return errs.ErrTransactionNotFound
	}

	if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		deletedRows, err = sess.ID(oldTransaction.RelatedId).Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=?", uid, false).Update(updateModel)

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrTransactionNotFound
		}
	}

	if _, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", uid, false, oldTransaction.TransactionId).Update(tagIndexUpdateModel); err != nil {
		return err
	}

	if _, err := sess.Cols("deleted", "deleted_unix_time").Where("uid=? AND deleted=? AND transaction_id=?", uid, false, oldTransaction.TransactionId).Update(pictureUpdateModel); err != nil {
		return err
	}

	sourceAccount.UpdatedUnixTime = now

	switch oldTransaction.Type {
	case models.TRANSACTION_DB_TYPE_MODIFY_BALANCE:
		if oldTransaction.RelatedAccountAmount != 0 {
			if _, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", oldTransaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount); err != nil {
				return err
			}
		}
	case models.TRANSACTION_DB_TYPE_INCOME:
		if oldTransaction.Amount != 0 {
			if _, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", oldTransaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount); err != nil {
				return err
			}
		}
	case models.TRANSACTION_DB_TYPE_EXPENSE:
		if oldTransaction.Amount != 0 {
			if _, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", oldTransaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount); err != nil {
				return err
			}
		}
	case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
		if oldTransaction.Amount != 0 {
			if _, err := sess.ID(sourceAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance+(%d)", oldTransaction.Amount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", sourceAccount.Uid, false).Update(sourceAccount); err != nil {
				return err
			}
		}

		if oldTransaction.RelatedAccountAmount != 0 {
			destinationAccount.UpdatedUnixTime = now

			if _, err := sess.ID(destinationAccount.AccountId).SetExpr("balance", fmt.Sprintf("balance-(%d)", oldTransaction.RelatedAccountAmount)).Cols("updated_unix_time").Where("uid=? AND deleted=?", destinationAccount.Uid, false).Update(destinationAccount); err != nil {
				return err
			}
		}
	}

	log.Infof(c, "[installments.deleteTransactionInSession] deleted transaction \"id:%d\" successfully", transactionId)
	return nil
}
