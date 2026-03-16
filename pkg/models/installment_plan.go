package models

import (
	"sort"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
)

// Installment provider keys.
const (
	INSTALLMENT_PROVIDER_HUABEI     = "huabei"
	INSTALLMENT_PROVIDER_JD_BAITIAO = "jd_baitiao"
	INSTALLMENT_PROVIDER_CUSTOM     = "custom"
)

// InstallmentAccountingMode represents installment accounting mode.
type InstallmentAccountingMode byte

// Installment accounting modes.
const (
	INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED  InstallmentAccountingMode = 1
	INSTALLMENT_ACCOUNTING_MODE_REPAYMENT_RECOGNIZED InstallmentAccountingMode = 2
)

// InstallmentDueDateSource represents due date source.
type InstallmentDueDateSource byte

// Installment due date sources.
const (
	INSTALLMENT_DUE_DATE_SOURCE_PLAN_RULE    InstallmentDueDateSource = 1
	INSTALLMENT_DUE_DATE_SOURCE_ACCOUNT_RULE InstallmentDueDateSource = 2
)

// InstallmentStorageMode represents future repayment storage mode.
type InstallmentStorageMode byte

// Installment storage modes.
const (
	INSTALLMENT_STORAGE_MODE_PLAN_ITEMS_ONLY          InstallmentStorageMode = 1
	INSTALLMENT_STORAGE_MODE_GENERATED_SCHEDULE_ITEMS InstallmentStorageMode = 2
)

// InstallmentItemStatus represents installment item status.
type InstallmentItemStatus byte

// Installment item statuses.
const (
	INSTALLMENT_ITEM_STATUS_UNPAID InstallmentItemStatus = 1
	INSTALLMENT_ITEM_STATUS_PAID   InstallmentItemStatus = 2
)

// InstallmentPlan represents a user installment plan.
type InstallmentPlan struct {
	PlanId                       int64                     `xorm:"PK"`
	Uid                          int64                     `xorm:"INDEX(IDX_installment_plan_uid_deleted_purchase_time) NOT NULL"`
	Deleted                      bool                      `xorm:"INDEX(IDX_installment_plan_uid_deleted_purchase_time) NOT NULL"`
	ProviderKey                  string                    `xorm:"VARCHAR(32) NOT NULL"`
	CustomProviderName           string                    `xorm:"VARCHAR(64) NOT NULL"`
	LiabilityAccountId           int64                     `xorm:"NOT NULL"`
	DefaultPaymentAccountId      int64                     `xorm:"NOT NULL"`
	AccountingMode               InstallmentAccountingMode `xorm:"NOT NULL"`
	PurchaseTransactionId        int64                     `xorm:"NOT NULL"`
	GeneratedPurchaseTransaction bool                      `xorm:"NOT NULL"`
	PurchaseCategoryId           int64                     `xorm:"NOT NULL"`
	RepaymentCategoryId          int64                     `xorm:"NOT NULL"`
	FeeCategoryId                int64                     `xorm:"NOT NULL"`
	TransferCategoryId           int64                     `xorm:"NOT NULL"`
	Title                        string                    `xorm:"VARCHAR(64) NOT NULL"`
	Notes                        string                    `xorm:"VARCHAR(255) NOT NULL"`
	PurchaseTime                 int64                     `xorm:"INDEX(IDX_installment_plan_uid_deleted_purchase_time) NOT NULL"`
	PurchaseTimezoneUtcOffset    int16                     `xorm:"NOT NULL"`
	Currency                     string                    `xorm:"VARCHAR(3) NOT NULL"`
	PrincipalTotal               int64                     `xorm:"NOT NULL"`
	FeeTotal                     int64                     `xorm:"NOT NULL"`
	PeriodCount                  int16                     `xorm:"NOT NULL"`
	DueDateSource                InstallmentDueDateSource  `xorm:"NOT NULL"`
	StorageMode                  InstallmentStorageMode    `xorm:"NOT NULL"`
	FirstDueDate                 string                    `xorm:"VARCHAR(10) NOT NULL"`
	MonthlyDueDay                int8                      `xorm:"NOT NULL"`
	CreatedUnixTime              int64
	UpdatedUnixTime              int64
	DeletedUnixTime              int64
}

// InstallmentItem represents an installment repayment item.
type InstallmentItem struct {
	ItemId                 int64                 `xorm:"PK"`
	PlanId                 int64                 `xorm:"INDEX(IDX_installment_item_plan_id_deleted_seq_no) NOT NULL"`
	Deleted                bool                  `xorm:"INDEX(IDX_installment_item_plan_id_deleted_seq_no) INDEX(IDX_installment_item_status_due_date) NOT NULL"`
	SeqNo                  int16                 `xorm:"INDEX(IDX_installment_item_plan_id_deleted_seq_no) NOT NULL"`
	DueDate                string                `xorm:"INDEX(IDX_installment_item_status_due_date) VARCHAR(10) NOT NULL"`
	PrincipalAmount        int64                 `xorm:"NOT NULL"`
	FeeAmount              int64                 `xorm:"NOT NULL"`
	DueAmount              int64                 `xorm:"NOT NULL"`
	Status                 InstallmentItemStatus `xorm:"INDEX(IDX_installment_item_status_due_date) NOT NULL"`
	PaidTime               int64                 `xorm:"NOT NULL"`
	ExpenseTransactionId   int64                 `xorm:"NOT NULL"`
	RepaymentTransactionId int64                 `xorm:"NOT NULL"`
	FeeTransactionId       int64                 `xorm:"NOT NULL"`
	GeneratedTemplateId    int64                 `xorm:"NOT NULL"`
	CreatedUnixTime        int64
	UpdatedUnixTime        int64
	DeletedUnixTime        int64
}

// InstallmentAccountRule represents phase 2 account rule.
type InstallmentAccountRule struct {
	RuleId             int64  `xorm:"PK"`
	Uid                int64  `xorm:"INDEX(IDX_installment_account_rule_uid_deleted) NOT NULL"`
	Deleted            bool   `xorm:"INDEX(IDX_installment_account_rule_uid_deleted) NOT NULL"`
	LiabilityAccountId int64  `xorm:"NOT NULL"`
	StatementDay       int8   `xorm:"NOT NULL"`
	RepaymentDay       int8   `xorm:"NOT NULL"`
	Timezone           string `xorm:"VARCHAR(64) NOT NULL"`
	Enabled            bool   `xorm:"NOT NULL"`
	CreatedUnixTime    int64
	UpdatedUnixTime    int64
	DeletedUnixTime    int64
}

// InstallmentAccountRuleListRequest represents list request.
type InstallmentAccountRuleListRequest struct{}

// InstallmentAccountRuleGetRequest represents get request.
type InstallmentAccountRuleGetRequest struct {
	LiabilityAccountId int64 `form:"liabilityAccountId,string" binding:"required,min=1"`
}

// InstallmentAccountRuleSaveRequest represents save request.
type InstallmentAccountRuleSaveRequest struct {
	LiabilityAccountId int64  `json:"liabilityAccountId,string" binding:"required,min=1"`
	StatementDay       int8   `json:"statementDay" binding:"required,min=1,max=31"`
	RepaymentDay       int8   `json:"repaymentDay" binding:"required,min=1,max=31"`
	Timezone           string `json:"timezone" binding:"required,max=64"`
	Enabled            bool   `json:"enabled"`
}

// InstallmentAccountRuleDeleteRequest represents delete request.
type InstallmentAccountRuleDeleteRequest struct {
	LiabilityAccountId int64 `json:"liabilityAccountId,string" binding:"required,min=1"`
}

// InstallmentPlanListRequest represents list request.
type InstallmentPlanListRequest struct{}

// InstallmentPlanGetRequest represents get request.
type InstallmentPlanGetRequest struct {
	Id int64 `form:"id,string" binding:"required,min=1"`
}

// InstallmentItemUpsertRequest represents item request body.
type InstallmentItemUpsertRequest struct {
	SeqNo           int16  `json:"seqNo" binding:"required,min=1"`
	DueDate         string `json:"dueDate" binding:"required,len=10"`
	PrincipalAmount int64  `json:"principalAmount" binding:"min=0,max=99999999999"`
	FeeAmount       int64  `json:"feeAmount" binding:"min=0,max=99999999999"`
	DueAmount       int64  `json:"dueAmount" binding:"min=0,max=99999999999"`
}

// InstallmentPlanCreateRequest represents create request.
type InstallmentPlanCreateRequest struct {
	ProviderKey                 string                          `json:"providerKey" binding:"required,max=32"`
	CustomProviderName          string                          `json:"customProviderName" binding:"max=64"`
	LiabilityAccountId          int64                           `json:"liabilityAccountId,string" binding:"required,min=1"`
	DefaultPaymentAccountId     int64                           `json:"defaultPaymentAccountId,string" binding:"min=0"`
	AccountingMode              InstallmentAccountingMode       `json:"accountingMode" binding:"required"`
	PurchaseTransactionId       int64                           `json:"purchaseTransactionId,string" binding:"min=0"`
	GeneratePurchaseTransaction bool                            `json:"generatePurchaseTransaction"`
	PurchaseCategoryId          int64                           `json:"purchaseCategoryId,string" binding:"min=0"`
	RepaymentCategoryId         int64                           `json:"repaymentCategoryId,string" binding:"min=0"`
	FeeCategoryId               int64                           `json:"feeCategoryId,string" binding:"min=0"`
	TransferCategoryId          int64                           `json:"transferCategoryId,string" binding:"min=0"`
	Title                       string                          `json:"title" binding:"required,notBlank,max=64"`
	Notes                       string                          `json:"notes" binding:"max=255"`
	PurchaseTime                int64                           `json:"purchaseTime" binding:"required,min=1"`
	PurchaseUtcOffset           int16                           `json:"purchaseUtcOffset" binding:"min=-720,max=840"`
	Currency                    string                          `json:"currency" binding:"required,len=3,validCurrency"`
	PrincipalTotal              int64                           `json:"principalTotal" binding:"required,min=1,max=99999999999"`
	FeeTotal                    int64                           `json:"feeTotal" binding:"min=0,max=99999999999"`
	PeriodCount                 int16                           `json:"periodCount" binding:"required,min=1,max=120"`
	DueDateSource               InstallmentDueDateSource        `json:"dueDateSource" binding:"required"`
	StorageMode                 InstallmentStorageMode          `json:"storageMode" binding:"required"`
	FirstDueDate                string                          `json:"firstDueDate" binding:"required,len=10"`
	MonthlyDueDay               int8                            `json:"monthlyDueDay" binding:"required,min=1,max=31"`
	Items                       []*InstallmentItemUpsertRequest `json:"items" binding:"required,min=1,max=120"`
}

// InstallmentPlanModifyRequest represents modify request.
type InstallmentPlanModifyRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
	InstallmentPlanCreateRequest
}

// InstallmentPlanDeleteRequest represents delete request.
type InstallmentPlanDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// InstallmentItemPayRequest represents item pay request.
type InstallmentItemPayRequest struct {
	ItemId              int64  `json:"itemId,string" binding:"required,min=1"`
	PaymentAccountId    int64  `json:"paymentAccountId,string" binding:"min=0"`
	Time                int64  `json:"time" binding:"required,min=1"`
	UtcOffset           int16  `json:"utcOffset" binding:"min=-720,max=840"`
	TransferCategoryId  int64  `json:"transferCategoryId,string" binding:"min=0"`
	RepaymentCategoryId int64  `json:"repaymentCategoryId,string" binding:"min=0"`
	FeeCategoryId       int64  `json:"feeCategoryId,string" binding:"min=0"`
	Comment             string `json:"comment" binding:"max=255"`
}

// InstallmentItemUnpayRequest represents item unpay request.
type InstallmentItemUnpayRequest struct {
	ItemId int64 `json:"itemId,string" binding:"required,min=1"`
}

// InstallmentItemInfoResponse represents item response.
type InstallmentItemInfoResponse struct {
	Id                     int64                 `json:"id,string"`
	SeqNo                  int16                 `json:"seqNo"`
	DueDate                string                `json:"dueDate"`
	PrincipalAmount        int64                 `json:"principalAmount"`
	FeeAmount              int64                 `json:"feeAmount"`
	DueAmount              int64                 `json:"dueAmount"`
	Status                 InstallmentItemStatus `json:"status"`
	PaidTime               int64                 `json:"paidTime"`
	ExpenseTransactionId   int64                 `json:"expenseTransactionId,string,omitempty"`
	RepaymentTransactionId int64                 `json:"repaymentTransactionId,string,omitempty"`
	FeeTransactionId       int64                 `json:"feeTransactionId,string,omitempty"`
	GeneratedTemplateId    int64                 `json:"generatedTemplateId,string,omitempty"`
}

// InstallmentPlanInfoResponse represents plan response.
type InstallmentPlanInfoResponse struct {
	Id                           int64                            `json:"id,string"`
	ProviderKey                  string                           `json:"providerKey"`
	ProviderName                 string                           `json:"providerName"`
	CustomProviderName           string                           `json:"customProviderName,omitempty"`
	LiabilityAccountId           int64                            `json:"liabilityAccountId,string"`
	DefaultPaymentAccountId      int64                            `json:"defaultPaymentAccountId,string,omitempty"`
	AccountingMode               InstallmentAccountingMode        `json:"accountingMode"`
	PurchaseTransactionId        int64                            `json:"purchaseTransactionId,string,omitempty"`
	GeneratedPurchaseTransaction bool                             `json:"generatedPurchaseTransaction"`
	PurchaseCategoryId           int64                            `json:"purchaseCategoryId,string,omitempty"`
	RepaymentCategoryId          int64                            `json:"repaymentCategoryId,string,omitempty"`
	FeeCategoryId                int64                            `json:"feeCategoryId,string,omitempty"`
	TransferCategoryId           int64                            `json:"transferCategoryId,string,omitempty"`
	Title                        string                           `json:"title"`
	Notes                        string                           `json:"notes"`
	PurchaseTime                 int64                            `json:"purchaseTime"`
	PurchaseUtcOffset            int16                            `json:"purchaseUtcOffset"`
	Currency                     string                           `json:"currency"`
	PrincipalTotal               int64                            `json:"principalTotal"`
	FeeTotal                     int64                            `json:"feeTotal"`
	DueTotal                     int64                            `json:"dueTotal"`
	PeriodCount                  int16                            `json:"periodCount"`
	DueDateSource                InstallmentDueDateSource         `json:"dueDateSource"`
	StorageMode                  InstallmentStorageMode           `json:"storageMode"`
	FirstDueDate                 string                           `json:"firstDueDate"`
	MonthlyDueDay                int8                             `json:"monthlyDueDay"`
	PaidCount                    int16                            `json:"paidCount"`
	UnpaidCount                  int16                            `json:"unpaidCount"`
	OverdueCount                 int16                            `json:"overdueCount"`
	NextUnpaidItem               *InstallmentItemInfoResponse     `json:"nextUnpaidItem,omitempty"`
	Items                        InstallmentItemInfoResponseSlice `json:"items,omitempty"`
}

// InstallmentAccountRuleInfoResponse represents account rule response.
type InstallmentAccountRuleInfoResponse struct {
	Id                 int64  `json:"id,string"`
	LiabilityAccountId int64  `json:"liabilityAccountId,string"`
	StatementDay       int8   `json:"statementDay"`
	RepaymentDay       int8   `json:"repaymentDay"`
	Timezone           string `json:"timezone"`
	Enabled            bool   `json:"enabled"`
}

// InstallmentAccountRuleInfoResponseSlice represents account rule response slice.
type InstallmentAccountRuleInfoResponseSlice []*InstallmentAccountRuleInfoResponse

// InstallmentItemInfoResponseSlice represents item response slice.
type InstallmentItemInfoResponseSlice []*InstallmentItemInfoResponse

// Len returns item count.
func (s InstallmentItemInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two items.
func (s InstallmentItemInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first item should be sorted before the second.
func (s InstallmentItemInfoResponseSlice) Less(i, j int) bool {
	return s[i].SeqNo < s[j].SeqNo
}

// InstallmentPlanInfoResponseSlice represents plan response slice.
type InstallmentPlanInfoResponseSlice []*InstallmentPlanInfoResponse

// Len returns plan count.
func (s InstallmentPlanInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two plans.
func (s InstallmentPlanInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first plan should be sorted before the second.
func (s InstallmentPlanInfoResponseSlice) Less(i, j int) bool {
	if s[i].PurchaseTime != s[j].PurchaseTime {
		return s[i].PurchaseTime > s[j].PurchaseTime
	}

	return s[i].Id > s[j].Id
}

// Len returns rule count.
func (s InstallmentAccountRuleInfoResponseSlice) Len() int {
	return len(s)
}

// Swap swaps two rules.
func (s InstallmentAccountRuleInfoResponseSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less reports whether the first rule should be sorted before the second.
func (s InstallmentAccountRuleInfoResponseSlice) Less(i, j int) bool {
	if s[i].LiabilityAccountId != s[j].LiabilityAccountId {
		return s[i].LiabilityAccountId < s[j].LiabilityAccountId
	}

	return s[i].Id < s[j].Id
}

// IsInstallmentProviderKeyValid returns whether provider key is supported.
func IsInstallmentProviderKeyValid(providerKey string) bool {
	return providerKey == INSTALLMENT_PROVIDER_HUABEI ||
		providerKey == INSTALLMENT_PROVIDER_JD_BAITIAO ||
		providerKey == INSTALLMENT_PROVIDER_CUSTOM
}

// GetInstallmentProviderName returns provider name.
func GetInstallmentProviderName(providerKey string, customProviderName string) string {
	switch providerKey {
	case INSTALLMENT_PROVIDER_HUABEI:
		return "花呗"
	case INSTALLMENT_PROVIDER_JD_BAITIAO:
		return "京东白条"
	case INSTALLMENT_PROVIDER_CUSTOM:
		return customProviderName
	default:
		return ""
	}
}

// Validate validates the plan values which do not require database lookups.
func (p *InstallmentPlan) Validate(items []*InstallmentItem) error {
	if !IsInstallmentProviderKeyValid(p.ProviderKey) || (p.ProviderKey == INSTALLMENT_PROVIDER_CUSTOM && p.CustomProviderName == "") {
		return errs.ErrInstallmentProviderInvalid
	}

	if p.AccountingMode != INSTALLMENT_ACCOUNTING_MODE_PURCHASE_RECOGNIZED &&
		p.AccountingMode != INSTALLMENT_ACCOUNTING_MODE_REPAYMENT_RECOGNIZED {
		return errs.ErrInstallmentAccountingModeInvalid
	}

	if p.DueDateSource != INSTALLMENT_DUE_DATE_SOURCE_PLAN_RULE &&
		p.DueDateSource != INSTALLMENT_DUE_DATE_SOURCE_ACCOUNT_RULE {
		return errs.ErrInstallmentDueDateSourceInvalid
	}

	if p.StorageMode != INSTALLMENT_STORAGE_MODE_PLAN_ITEMS_ONLY &&
		p.StorageMode != INSTALLMENT_STORAGE_MODE_GENERATED_SCHEDULE_ITEMS {
		return errs.ErrInstallmentStorageModeInvalid
	}

	if p.PeriodCount <= 0 || len(items) != int(p.PeriodCount) {
		return errs.ErrInstallmentPeriodCountInvalid
	}

	if err := ValidateInstallmentDate(p.FirstDueDate); err != nil {
		return err
	}

	if p.MonthlyDueDay <= 0 || p.MonthlyDueDay > 31 {
		return errs.ErrInstallmentDateInvalid
	}

	return ValidateInstallmentItems(items, p.PrincipalTotal, p.FeeTotal)
}

// Validate validates the account rule values which do not require database lookups.
func (r *InstallmentAccountRule) Validate() error {
	if r.LiabilityAccountId <= 0 || r.StatementDay <= 0 || r.StatementDay > 31 || r.RepaymentDay <= 0 || r.RepaymentDay > 31 {
		return errs.ErrInstallmentAccountRuleInvalid
	}

	if r.Timezone == "" {
		return errs.ErrInstallmentTimezoneInvalid
	}

	location, err := time.LoadLocation(r.Timezone)

	if err != nil || location == nil {
		return errs.ErrInstallmentTimezoneInvalid
	}

	return nil
}

// ValidateInstallmentDate validates date string.
func ValidateInstallmentDate(date string) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return errs.ErrInstallmentDateInvalid
	}

	return nil
}

// ValidateInstallmentItems validates all installment items.
func ValidateInstallmentItems(items []*InstallmentItem, principalTotal int64, feeTotal int64) error {
	if len(items) < 1 {
		return errs.ErrInstallmentItemsInvalid
	}

	totalPrincipal := int64(0)
	totalFee := int64(0)
	seqNoMap := make(map[int16]bool, len(items))

	for i := 0; i < len(items); i++ {
		item := items[i]

		if item.SeqNo <= 0 {
			return errs.ErrInstallmentItemsInvalid
		}

		if seqNoMap[item.SeqNo] {
			return errs.ErrInstallmentDuplicateItemSequence
		}

		seqNoMap[item.SeqNo] = true

		if err := ValidateInstallmentDate(item.DueDate); err != nil {
			return err
		}

		if item.PrincipalAmount < 0 || item.FeeAmount < 0 || item.DueAmount < 0 || item.DueAmount != item.PrincipalAmount+item.FeeAmount {
			return errs.ErrInstallmentAmountMismatch
		}

		totalPrincipal += item.PrincipalAmount
		totalFee += item.FeeAmount
	}

	if totalPrincipal != principalTotal || totalFee != feeTotal {
		return errs.ErrInstallmentAmountMismatch
	}

	return nil
}

// ToInstallmentItemInfoResponse converts item to response.
func (i *InstallmentItem) ToInstallmentItemInfoResponse() *InstallmentItemInfoResponse {
	return &InstallmentItemInfoResponse{
		Id:                     i.ItemId,
		SeqNo:                  i.SeqNo,
		DueDate:                i.DueDate,
		PrincipalAmount:        i.PrincipalAmount,
		FeeAmount:              i.FeeAmount,
		DueAmount:              i.DueAmount,
		Status:                 i.Status,
		PaidTime:               i.PaidTime,
		ExpenseTransactionId:   i.ExpenseTransactionId,
		RepaymentTransactionId: i.RepaymentTransactionId,
		FeeTransactionId:       i.FeeTransactionId,
		GeneratedTemplateId:    i.GeneratedTemplateId,
	}
}

// ToInstallmentAccountRuleInfoResponse converts rule to response.
func (r *InstallmentAccountRule) ToInstallmentAccountRuleInfoResponse() *InstallmentAccountRuleInfoResponse {
	return &InstallmentAccountRuleInfoResponse{
		Id:                 r.RuleId,
		LiabilityAccountId: r.LiabilityAccountId,
		StatementDay:       r.StatementDay,
		RepaymentDay:       r.RepaymentDay,
		Timezone:           r.Timezone,
		Enabled:            r.Enabled,
	}
}

// ToInstallmentPlanInfoResponse converts plan to response.
func (p *InstallmentPlan) ToInstallmentPlanInfoResponse(items []*InstallmentItem, withItems bool) *InstallmentPlanInfoResponse {
	resp := &InstallmentPlanInfoResponse{
		Id:                           p.PlanId,
		ProviderKey:                  p.ProviderKey,
		ProviderName:                 GetInstallmentProviderName(p.ProviderKey, p.CustomProviderName),
		CustomProviderName:           p.CustomProviderName,
		LiabilityAccountId:           p.LiabilityAccountId,
		DefaultPaymentAccountId:      p.DefaultPaymentAccountId,
		AccountingMode:               p.AccountingMode,
		PurchaseTransactionId:        p.PurchaseTransactionId,
		GeneratedPurchaseTransaction: p.GeneratedPurchaseTransaction,
		PurchaseCategoryId:           p.PurchaseCategoryId,
		RepaymentCategoryId:          p.RepaymentCategoryId,
		FeeCategoryId:                p.FeeCategoryId,
		TransferCategoryId:           p.TransferCategoryId,
		Title:                        p.Title,
		Notes:                        p.Notes,
		PurchaseTime:                 p.PurchaseTime,
		PurchaseUtcOffset:            p.PurchaseTimezoneUtcOffset,
		Currency:                     p.Currency,
		PrincipalTotal:               p.PrincipalTotal,
		FeeTotal:                     p.FeeTotal,
		DueTotal:                     p.PrincipalTotal + p.FeeTotal,
		PeriodCount:                  p.PeriodCount,
		DueDateSource:                p.DueDateSource,
		StorageMode:                  p.StorageMode,
		FirstDueDate:                 p.FirstDueDate,
		MonthlyDueDay:                p.MonthlyDueDay,
	}

	if len(items) < 1 {
		return resp
	}

	today := time.Now().Format("2006-01-02")
	itemResps := make(InstallmentItemInfoResponseSlice, len(items))

	for i := 0; i < len(items); i++ {
		itemResp := items[i].ToInstallmentItemInfoResponse()
		itemResps[i] = itemResp

		if items[i].Status == INSTALLMENT_ITEM_STATUS_PAID {
			resp.PaidCount++
			continue
		}

		resp.UnpaidCount++

		if items[i].DueDate < today {
			resp.OverdueCount++
		}

		if resp.NextUnpaidItem == nil || items[i].SeqNo < resp.NextUnpaidItem.SeqNo {
			resp.NextUnpaidItem = itemResp
		}
	}

	sort.Sort(itemResps)

	if withItems {
		resp.Items = itemResps
	}

	return resp
}
