package api

import (
	"sort"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// InstallmentsApi represents installment api.
type InstallmentsApi struct {
	ApiUsingConfig
	installments *services.InstallmentService
}

// Initialize an installment api singleton instance.
var (
	InstallmentPlans = &InstallmentsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		installments: services.Installments,
	}
)

// PlanListHandler returns installment plan list of current user.
func (a *InstallmentsApi) PlanListHandler(c *core.WebContext) (any, *errs.Error) {
	var planListReq models.InstallmentPlanListRequest
	err := c.ShouldBindQuery(&planListReq)

	if err != nil {
		log.Warnf(c, "[installments.PlanListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	plans, err := a.installments.GetAllPlansByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[installments.PlanListHandler] failed to get installment plans for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	planIds := make([]int64, len(plans))

	for i := 0; i < len(plans); i++ {
		planIds[i] = plans[i].PlanId
	}

	itemMap, err := a.installments.GetItemsByPlanIds(c, uid, planIds)

	if err != nil {
		log.Errorf(c, "[installments.PlanListHandler] failed to get installment items for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	planResps := make(models.InstallmentPlanInfoResponseSlice, len(plans))

	for i := 0; i < len(plans); i++ {
		planResps[i] = plans[i].ToInstallmentPlanInfoResponse(itemMap[plans[i].PlanId], false)
	}

	sort.Sort(planResps)

	return planResps, nil
}

// PlanGetHandler returns one specific installment plan of current user.
func (a *InstallmentsApi) PlanGetHandler(c *core.WebContext) (any, *errs.Error) {
	var planGetReq models.InstallmentPlanGetRequest
	err := c.ShouldBindQuery(&planGetReq)

	if err != nil {
		log.Warnf(c, "[installments.PlanGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	return a.getPlanResponse(c, c.GetCurrentUid(), planGetReq.Id)
}

// PlanCreateHandler saves a new installment plan.
func (a *InstallmentsApi) PlanCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var planCreateReq models.InstallmentPlanCreateRequest
	err := c.ShouldBindJSON(&planCreateReq)

	if err != nil {
		log.Warnf(c, "[installments.PlanCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	plan, items := a.createPlanModel(uid, &planCreateReq)
	err = a.installments.CreatePlan(c, plan, items)

	if err != nil {
		log.Errorf(c, "[installments.PlanCreateHandler] failed to create installment plan for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[installments.PlanCreateHandler] user \"uid:%d\" has created installment plan \"id:%d\" successfully", uid, plan.PlanId)
	return a.getPlanResponse(c, uid, plan.PlanId)
}

// PlanModifyHandler updates an installment plan.
func (a *InstallmentsApi) PlanModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var planModifyReq models.InstallmentPlanModifyRequest
	err := c.ShouldBindJSON(&planModifyReq)

	if err != nil {
		log.Warnf(c, "[installments.PlanModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	plan, items := a.createPlanModel(uid, &planModifyReq.InstallmentPlanCreateRequest)
	plan.PlanId = planModifyReq.Id
	err = a.installments.ModifyPlan(c, plan, items)

	if err != nil {
		log.Errorf(c, "[installments.PlanModifyHandler] failed to modify installment plan \"id:%d\" for user \"uid:%d\", because %s", planModifyReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[installments.PlanModifyHandler] user \"uid:%d\" has modified installment plan \"id:%d\" successfully", uid, planModifyReq.Id)
	return a.getPlanResponse(c, uid, planModifyReq.Id)
}

// PlanDeleteHandler deletes an installment plan.
func (a *InstallmentsApi) PlanDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var planDeleteReq models.InstallmentPlanDeleteRequest
	err := c.ShouldBindJSON(&planDeleteReq)

	if err != nil {
		log.Warnf(c, "[installments.PlanDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.installments.DeletePlan(c, uid, planDeleteReq.Id)

	if err != nil {
		log.Errorf(c, "[installments.PlanDeleteHandler] failed to delete installment plan \"id:%d\" for user \"uid:%d\", because %s", planDeleteReq.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[installments.PlanDeleteHandler] user \"uid:%d\" has deleted installment plan \"id:%d\"", uid, planDeleteReq.Id)
	return true, nil
}

// PayItemHandler pays one installment item.
func (a *InstallmentsApi) PayItemHandler(c *core.WebContext) (any, *errs.Error) {
	var payReq models.InstallmentItemPayRequest
	err := c.ShouldBindJSON(&payReq)

	if err != nil {
		log.Warnf(c, "[installments.PayItemHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.installments.PayItem(c, uid, &payReq)

	if err != nil {
		log.Errorf(c, "[installments.PayItemHandler] failed to pay installment item \"id:%d\" for user \"uid:%d\", because %s", payReq.ItemId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[installments.PayItemHandler] user \"uid:%d\" has paid installment item \"id:%d\"", uid, payReq.ItemId)
	return true, nil
}

// UnpayItemHandler reverts one installment item payment.
func (a *InstallmentsApi) UnpayItemHandler(c *core.WebContext) (any, *errs.Error) {
	var unpayReq models.InstallmentItemUnpayRequest
	err := c.ShouldBindJSON(&unpayReq)

	if err != nil {
		log.Warnf(c, "[installments.UnpayItemHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.installments.UnpayItem(c, uid, unpayReq.ItemId)

	if err != nil {
		log.Errorf(c, "[installments.UnpayItemHandler] failed to unpay installment item \"id:%d\" for user \"uid:%d\", because %s", unpayReq.ItemId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[installments.UnpayItemHandler] user \"uid:%d\" has unpaid installment item \"id:%d\"", uid, unpayReq.ItemId)
	return true, nil
}

// AccountRuleListHandler returns installment account rule list of current user.
func (a *InstallmentsApi) AccountRuleListHandler(c *core.WebContext) (any, *errs.Error) {
	var listReq models.InstallmentAccountRuleListRequest
	err := c.ShouldBindQuery(&listReq)

	if err != nil {
		log.Warnf(c, "[installments.AccountRuleListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	rules, err := a.installments.GetAllAccountRulesByUid(c, uid)

	if err != nil {
		log.Errorf(c, "[installments.AccountRuleListHandler] failed to get installment account rules for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	resps := make(models.InstallmentAccountRuleInfoResponseSlice, len(rules))

	for i := 0; i < len(rules); i++ {
		resps[i] = rules[i].ToInstallmentAccountRuleInfoResponse()
	}

	sort.Sort(resps)
	return resps, nil
}

// AccountRuleGetHandler returns one installment account rule of current user.
func (a *InstallmentsApi) AccountRuleGetHandler(c *core.WebContext) (any, *errs.Error) {
	var getReq models.InstallmentAccountRuleGetRequest
	err := c.ShouldBindQuery(&getReq)

	if err != nil {
		log.Warnf(c, "[installments.AccountRuleGetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	rule, err := a.installments.GetAccountRuleByLiabilityAccountId(c, uid, getReq.LiabilityAccountId)

	if err != nil {
		log.Errorf(c, "[installments.AccountRuleGetHandler] failed to get installment account rule of liability account \"id:%d\" for user \"uid:%d\", because %s", getReq.LiabilityAccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return rule.ToInstallmentAccountRuleInfoResponse(), nil
}

// AccountRuleSaveHandler saves an installment account rule.
func (a *InstallmentsApi) AccountRuleSaveHandler(c *core.WebContext) (any, *errs.Error) {
	var saveReq models.InstallmentAccountRuleSaveRequest
	err := c.ShouldBindJSON(&saveReq)

	if err != nil {
		log.Warnf(c, "[installments.AccountRuleSaveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	rule := &models.InstallmentAccountRule{
		Uid:                uid,
		LiabilityAccountId: saveReq.LiabilityAccountId,
		StatementDay:       saveReq.StatementDay,
		RepaymentDay:       saveReq.RepaymentDay,
		Timezone:           saveReq.Timezone,
		Enabled:            saveReq.Enabled,
	}
	err = a.installments.SaveAccountRule(c, rule)

	if err != nil {
		log.Errorf(c, "[installments.AccountRuleSaveHandler] failed to save installment account rule of liability account \"id:%d\" for user \"uid:%d\", because %s", saveReq.LiabilityAccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[installments.AccountRuleSaveHandler] user \"uid:%d\" has saved installment account rule of liability account \"id:%d\"", uid, saveReq.LiabilityAccountId)
	return rule.ToInstallmentAccountRuleInfoResponse(), nil
}

// AccountRuleDeleteHandler deletes an installment account rule.
func (a *InstallmentsApi) AccountRuleDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var deleteReq models.InstallmentAccountRuleDeleteRequest
	err := c.ShouldBindJSON(&deleteReq)

	if err != nil {
		log.Warnf(c, "[installments.AccountRuleDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.installments.DeleteAccountRule(c, uid, deleteReq.LiabilityAccountId)

	if err != nil {
		log.Errorf(c, "[installments.AccountRuleDeleteHandler] failed to delete installment account rule of liability account \"id:%d\" for user \"uid:%d\", because %s", deleteReq.LiabilityAccountId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[installments.AccountRuleDeleteHandler] user \"uid:%d\" has deleted installment account rule of liability account \"id:%d\"", uid, deleteReq.LiabilityAccountId)
	return true, nil
}

func (a *InstallmentsApi) getPlanResponse(c *core.WebContext, uid int64, planId int64) (any, *errs.Error) {
	plan, err := a.installments.GetPlanByPlanId(c, uid, planId)

	if err != nil {
		log.Errorf(c, "[installments.getPlanResponse] failed to get installment plan \"id:%d\" for user \"uid:%d\", because %s", planId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	items, err := a.installments.GetItemsByPlanId(c, uid, planId)

	if err != nil {
		log.Errorf(c, "[installments.getPlanResponse] failed to get installment items of plan \"id:%d\" for user \"uid:%d\", because %s", planId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return plan.ToInstallmentPlanInfoResponse(items, true), nil
}

func (a *InstallmentsApi) createPlanModel(uid int64, req *models.InstallmentPlanCreateRequest) (*models.InstallmentPlan, []*models.InstallmentItem) {
	items := make([]*models.InstallmentItem, len(req.Items))

	for i := 0; i < len(req.Items); i++ {
		items[i] = &models.InstallmentItem{
			SeqNo:           req.Items[i].SeqNo,
			DueDate:         req.Items[i].DueDate,
			PrincipalAmount: req.Items[i].PrincipalAmount,
			FeeAmount:       req.Items[i].FeeAmount,
			DueAmount:       req.Items[i].DueAmount,
		}
	}

	return &models.InstallmentPlan{
		Uid:                          uid,
		ProviderKey:                  req.ProviderKey,
		CustomProviderName:           req.CustomProviderName,
		LiabilityAccountId:           req.LiabilityAccountId,
		DefaultPaymentAccountId:      req.DefaultPaymentAccountId,
		AccountingMode:               req.AccountingMode,
		PurchaseTransactionId:        req.PurchaseTransactionId,
		GeneratedPurchaseTransaction: req.GeneratePurchaseTransaction,
		PurchaseCategoryId:           req.PurchaseCategoryId,
		RepaymentCategoryId:          req.RepaymentCategoryId,
		FeeCategoryId:                req.FeeCategoryId,
		TransferCategoryId:           req.TransferCategoryId,
		Title:                        req.Title,
		Notes:                        req.Notes,
		PurchaseTime:                 req.PurchaseTime,
		PurchaseTimezoneUtcOffset:    req.PurchaseUtcOffset,
		Currency:                     req.Currency,
		PrincipalTotal:               req.PrincipalTotal,
		FeeTotal:                     req.FeeTotal,
		PeriodCount:                  req.PeriodCount,
		DueDateSource:                req.DueDateSource,
		StorageMode:                  req.StorageMode,
		FirstDueDate:                 req.FirstDueDate,
		MonthlyDueDay:                req.MonthlyDueDay,
	}, items
}
