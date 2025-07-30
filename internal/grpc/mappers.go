package grpc

import (
	pb "ariand/gen/go/ariand/v1"
	"ariand/internal/domain"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// =======================
//
//	TO PROTO (Domain -> PB)
//
// =======================
func toProtoTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func toProtoAccountType(at domain.AccountType) pb.AccountType {
	switch at {
	case domain.AccountTypeChequing:
		return pb.AccountType_ACCOUNT_TYPE_CHEQUING
	case domain.AccountTypeSavings:
		return pb.AccountType_ACCOUNT_TYPE_SAVINGS
	case domain.AccountTypeCreditCard:
		return pb.AccountType_ACCOUNT_TYPE_CREDIT_CARD
	case domain.AccountTypeInvestment:
		return pb.AccountType_ACCOUNT_TYPE_INVESTMENT
	case domain.AccountTypeOther:
		return pb.AccountType_ACCOUNT_TYPE_OTHER
	default:
		return pb.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

func toProtoAccount(a *domain.Account) *pb.Account {
	if a == nil {
		return nil
	}
	return &pb.Account{
		Id:            a.ID,
		Name:          a.Name,
		Bank:          a.Bank,
		Type:          toProtoAccountType(a.Type),
		Alias:         a.Alias,
		AnchorDate:    toProtoTimestamp(a.AnchorDate),
		AnchorBalance: float64ToMoney(a.AnchorBalance, a.AnchorCurrency),
		CreatedAt:     toProtoTimestamp(a.CreatedAt),
		UpdatedAt:     toProtoTimestamp(a.UpdatedAt),
	}
}

func toProtoCategory(c *domain.Category) *pb.Category {
	if c == nil {
		return nil
	}
	return &pb.Category{
		Id:        c.ID,
		Slug:      c.Slug,
		Label:     c.Label,
		Color:     c.Color,
		CreatedAt: toProtoTimestamp(c.CreatedAt),
		UpdatedAt: toProtoTimestamp(c.UpdatedAt),
	}
}

func toProtoTransaction(t *domain.Transaction) *pb.Transaction {
	if t == nil {
		return nil
	}
	var balAfter *money.Money
	if t.BalanceAfter != nil {
		balAfter = float64ToMoney(*t.BalanceAfter, t.TxCurrency)
	}
	var fAmount *money.Money
	if t.ForeignAmount != nil && t.ForeignCurrency != nil {
		fAmount = float64ToMoney(*t.ForeignAmount, *t.ForeignCurrency)
	}

	return &pb.Transaction{
		Id:            t.ID,
		EmailId:       t.EmailID,
		AccountId:     t.AccountID,
		TxDate:        toProtoTimestamp(t.TxDate),
		TxAmount:      float64ToMoney(t.TxAmount, t.TxCurrency),
		Direction:     toProtoTransactionDirection(t.TxDirection),
		Description:   t.TxDesc,
		BalanceAfter:  balAfter,
		CategoryId:    t.CategoryID,
		CategorySlug:  t.CategorySlug,
		CatStatus:     toProtoCategorizationStatus(t.CatStatus),
		Merchant:      t.Merchant,
		UserNotes:     t.UserNotes,
		Suggestions:   []string(t.Suggestions),
		ReceiptId:     t.ReceiptID,
		ForeignAmount: fAmount,
		ExchangeRate:  t.ExchangeRate,
		CreatedAt:     toProtoTimestamp(t.CreatedAt),
		UpdatedAt:     toProtoTimestamp(t.UpdatedAt),
	}
}

func toProtoReceiptEngine(p domain.ReceiptProvider) pb.ReceiptEngine {
	switch p {
	case domain.ProviderGemini:
		return pb.ReceiptEngine_RECEIPT_ENGINE_GEMINI
	case domain.ProviderLocal:
		return pb.ReceiptEngine_RECEIPT_ENGINE_LOCAL
	default:
		return pb.ReceiptEngine_RECEIPT_ENGINE_UNSPECIFIED
	}
}

func toProtoReceipt(r *domain.Receipt) *pb.Receipt {
	if r == nil {
		return nil
	}

	items := make([]*pb.ReceiptItem, len(r.Items))
	for i, item := range r.Items {
		var lineNo int32
		if item.LineNo != nil {
			lineNo = int32(*item.LineNo)
		}

		currency := deref(r.Currency)

		items[i] = &pb.ReceiptItem{
			Id:           item.ID,
			ReceiptId:    item.ReceiptID,
			LineNo:       lineNo,
			Name:         item.Name,
			Quantity:     deref(item.Qty),
			UnitPrice:    float64ToMoney(deref(item.UnitPrice), currency),
			LineTotal:    float64ToMoney(deref(item.LineTotal), currency),
			Sku:          item.SKU,
			CategoryHint: item.CategoryHint,
			CreatedAt:    toProtoTimestamp(item.CreatedAt),
			UpdatedAt:    toProtoTimestamp(item.UpdatedAt),
		}
	}

	return &pb.Receipt{
		Id:            r.ID,
		TransactionId: r.TransactionID,
		Engine:        toProtoReceiptEngine(r.Provider),
		ParseStatus:   toProtoReceiptParseStatus(r.ParseStatus),
		LinkStatus:    toProtoReceiptLinkStatus(r.LinkStatus),
		MatchIds:      r.MatchSuggestions,
		Merchant:      r.Merchant,
		PurchaseDate:  timeToDate(deref(r.PurchaseDate)),
		TotalAmount:   float64ToMoney(deref(r.TotalAmount), deref(r.Currency)),
		TaxAmount:     float64ToMoney(deref(r.TaxAmount), deref(r.Currency)),
		RawPayload:    stringPtr(r.RawPayload),
		CanonicalData: stringPtr(r.CanonicalData),
		Items:         items,
		CreatedAt:     toProtoTimestamp(r.CreatedAt),
		UpdatedAt:     toProtoTimestamp(r.UpdatedAt),
	}
}

// =========================
//
//	FROM PROTO (PB -> Domain)
//
// =========================
func fromProtoTimestamp(ts *timestamppb.Timestamp) time.Time {
	if ts == nil || !ts.IsValid() {
		return time.Time{}
	}
	return ts.AsTime()
}

func fromProtoAccountType(at pb.AccountType) domain.AccountType {
	switch at {
	case pb.AccountType_ACCOUNT_TYPE_CHEQUING:
		return domain.AccountTypeChequing
	case pb.AccountType_ACCOUNT_TYPE_SAVINGS:
		return domain.AccountTypeSavings
	case pb.AccountType_ACCOUNT_TYPE_CREDIT_CARD:
		return domain.AccountTypeCreditCard
	case pb.AccountType_ACCOUNT_TYPE_INVESTMENT:
		return domain.AccountTypeInvestment
	case pb.AccountType_ACCOUNT_TYPE_OTHER:
		return domain.AccountTypeOther
	default:
		return ""
	}
}

func fromProtoCreateAccountRequest(req *pb.CreateAccountRequest) *domain.Account {
	balance, currency := moneyToFloat64(req.GetAnchorBalance())
	return &domain.Account{
		Name:           req.GetName(),
		Bank:           req.GetBank(),
		Type:           fromProtoAccountType(req.GetType()),
		Alias:          req.Alias,
		AnchorBalance:  balance,
		AnchorCurrency: currency,
		AnchorDate:     time.Now(),
	}
}

func fromProtoCreateTransactionRequest(req *pb.CreateTransactionRequest) *domain.Transaction {
	var fAmount *float64
	var fCurrency *string
	if req.GetForeignAmount() != nil {
		v, c := moneyToFloat64(req.GetForeignAmount())
		fAmount = &v
		fCurrency = &c
	}
	txDesc := req.GetTxDesc()
	txAmount, txCurrency := moneyToFloat64(req.GetTxAmount())

	return &domain.Transaction{
		EmailID:         req.EmailId,
		AccountID:       req.GetAccountId(),
		TxDate:          fromProtoTimestamp(req.GetTxDate()),
		TxAmount:        txAmount,
		TxCurrency:      txCurrency,
		TxDirection:     fromProtoTransactionDirection(req.GetTxDirection()),
		TxDesc:          &txDesc,
		CategoryID:      req.CategoryId,
		Merchant:        req.Merchant,
		UserNotes:       req.UserNotes,
		Suggestions:     pq.StringArray(req.Suggestions),
		ForeignAmount:   fAmount,
		ForeignCurrency: fCurrency,
		ExchangeRate:    req.ExchangeRate,
	}
}

func fromProtoReceiptEngine(p pb.ReceiptEngine) domain.ReceiptProvider {
	switch p {
	case pb.ReceiptEngine_RECEIPT_ENGINE_GEMINI:
		return domain.ProviderGemini
	case pb.ReceiptEngine_RECEIPT_ENGINE_LOCAL:
		return domain.ProviderLocal
	default:
		return domain.ProviderGemini // Default to Gemini
	}
}

// =======================
//
//	HELPERS
//
// =======================
func fieldsFromUpdateMask(mask *fieldmaskpb.FieldMask, in interface{}) map[string]any {
	fields := make(map[string]any)
	if mask == nil || in == nil {
		return fields
	}
	val := reflect.ValueOf(in)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for _, path := range mask.Paths {
		goFieldName := snakeToCamel(path)
		fieldVal := val.FieldByName(goFieldName)
		if fieldVal.IsValid() {
			fields[path] = fieldVal.Interface()
		}
	}
	return fields
}

func snakeToCamel(s string) string {
	var b strings.Builder
	capNext := true
	for _, r := range s {
		if r == '_' {
			capNext = true
			continue
		}
		if capNext {
			b.WriteRune(toUpper(r))
			capNext = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - ('a' - 'A')
	}
	return r
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

func stringPtr(jt types.JSONText) *string {
	if jt == nil {
		return nil
	}
	s := string(jt)
	return &s
}

func float64ToMoney(val float64, currency string) *money.Money {
	if currency == "" {
		currency = "XXX" // ISO 4217 code for "no currency"
	}
	units := int64(val)
	nanos := int32((val - float64(units)) * 1e9)
	return &money.Money{
		CurrencyCode: currency,
		Units:        units,
		Nanos:        nanos,
	}
}

func moneyToFloat64(m *money.Money) (float64, string) {
	if m == nil {
		return 0.0, ""
	}
	val := float64(m.Units) + float64(m.Nanos)/1e9
	return val, m.CurrencyCode
}

func timeToDate(t time.Time) *date.Date {
	if t.IsZero() {
		return nil
	}
	return &date.Date{
		Year:  int32(t.Year()),
		Month: int32(t.Month()),
		Day:   int32(t.Day()),
	}
}

func stringToDate(s string) (*date.Date, error) {
	if s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date string '%s': %w", s, err)
	}
	return timeToDate(t), nil
}

func toProtoCategorizationStatus(s domain.CategorizationStatus) pb.CategorizationStatus {
	switch s {
	case domain.CatStatusNone:
		return pb.CategorizationStatus_CATEGORIZATION_STATUS_NONE
	case domain.CatStatusAuto, domain.CatStatusAI:
		return pb.CategorizationStatus_CATEGORIZATION_STATUS_AUTO
	case domain.CatStatusManual:
		return pb.CategorizationStatus_CATEGORIZATION_STATUS_MANUAL
	case domain.CatStatusVerified:
		return pb.CategorizationStatus_CATEGORIZATION_STATUS_VERIFIED
	default:
		return pb.CategorizationStatus_CATEGORIZATION_STATUS_UNSPECIFIED
	}
}

func toProtoReceiptParseStatus(s domain.ReceiptParseStatus) pb.ReceiptParseStatus {
	switch s {
	case "pending":
		return pb.ReceiptParseStatus_RECEIPT_PARSE_STATUS_PENDING
	case "success":
		return pb.ReceiptParseStatus_RECEIPT_PARSE_STATUS_SUCCESS
	case "failed":
		return pb.ReceiptParseStatus_RECEIPT_PARSE_STATUS_FAILED
	default:
		return pb.ReceiptParseStatus_RECEIPT_PARSE_STATUS_UNSPECIFIED
	}
}

func toProtoReceiptLinkStatus(s domain.ReceiptLinkStatus) pb.ReceiptLinkStatus {
	switch s {
	case "unlinked":
		return pb.ReceiptLinkStatus_RECEIPT_LINK_STATUS_UNLINKED
	case "matched":
		return pb.ReceiptLinkStatus_RECEIPT_LINK_STATUS_MATCHED
	case "needs_verification":
		return pb.ReceiptLinkStatus_RECEIPT_LINK_STATUS_NEEDS_VERIFICATION
	default:
		return pb.ReceiptLinkStatus_RECEIPT_LINK_STATUS_UNSPECIFIED
	}
}

func toProtoTransactionDirection(s string) pb.TransactionDirection {
	switch s {
	case "in":
		return pb.TransactionDirection_TRANSACTION_DIRECTION_INCOMING
	case "out":
		return pb.TransactionDirection_TRANSACTION_DIRECTION_OUTGOING
	default:
		return pb.TransactionDirection_TRANSACTION_DIRECTION_UNSPECIFIED
	}
}

func fromProtoTransactionDirection(e pb.TransactionDirection) string {
	switch e {
	case pb.TransactionDirection_TRANSACTION_DIRECTION_INCOMING:
		return "in"
	case pb.TransactionDirection_TRANSACTION_DIRECTION_OUTGOING:
		return "out"
	default:
		return ""
	}
}
