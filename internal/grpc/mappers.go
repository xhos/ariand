package grpc

import (
	pb "ariand/gen/go/ariand/v1"
	"ariand/internal/domain"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// =======================
//  TO PROTO (Domain -> PB)
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
		AnchorBalance: a.AnchorBalance,
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
	return &pb.Transaction{
		Id:              t.ID,
		EmailId:         t.EmailID,
		AccountId:       t.AccountID,
		TxDate:          toProtoTimestamp(t.TxDate),
		TxAmount:        t.TxAmount,
		TxCurrency:      t.TxCurrency,
		TxDirection:     t.TxDirection,
		TxDesc:          t.TxDesc,
		BalanceAfter:    t.BalanceAfter,
		CategoryId:      t.CategoryID,
		CategorySlug:    t.CategorySlug,
		CategoryLabel:   t.CategoryLabel,
		CategoryColor:   t.CategoryColor,
		CatStatus:       t.CatStatus,
		Merchant:        t.Merchant,
		UserNotes:       t.UserNotes,
		Suggestions:     []string(t.Suggestions),
		ReceiptId:       t.ReceiptID,
		ForeignAmount:   t.ForeignAmount,
		ForeignCurrency: t.ForeignCurrency,
		ExchangeRate:    t.ExchangeRate,
		CreatedAt:       toProtoTimestamp(t.CreatedAt),
		UpdatedAt:       toProtoTimestamp(t.UpdatedAt),
	}
}

func toProtoReceiptProvider(p domain.ReceiptProvider) pb.ReceiptProvider {
	switch p {
	case domain.ProviderGemini:
		return pb.ReceiptProvider_RECEIPT_PROVIDER_GEMINI
	case domain.ProviderLocal:
		return pb.ReceiptProvider_RECEIPT_PROVIDER_LOCAL
	default:
		return pb.ReceiptProvider_RECEIPT_PROVIDER_UNSPECIFIED
	}
}

func toProtoReceipt(r *domain.Receipt) *pb.Receipt {
	if r == nil {
		return nil
	}

	items := make([]*pb.ReceiptItem, len(r.Items))
	for i, item := range r.Items {
		var lineNo *int32
		if item.LineNo != nil {
			v := int32(*item.LineNo)
			lineNo = &v
		}
		items[i] = &pb.ReceiptItem{
			Id:           item.ID,
			ReceiptId:    item.ReceiptID,
			LineNo:       lineNo,
			Name:         item.Name,
			Qty:          item.Qty,
			UnitPrice:    item.UnitPrice,
			LineTotal:    item.LineTotal,
			Sku:          item.SKU,
			CategoryHint: item.CategoryHint,
			CreatedAt:    toProtoTimestamp(item.CreatedAt),
			UpdatedAt:    toProtoTimestamp(item.UpdatedAt),
		}
	}

	return &pb.Receipt{
		Id:               r.ID,
		TransactionId:    r.TransactionID,
		Provider:         toProtoReceiptProvider(r.Provider),
		MatchSuggestions: r.MatchSuggestions,
		Merchant:         r.Merchant,
		PurchaseDate:     toProtoTimestamp(deref(r.PurchaseDate)),
		TotalAmount:      r.TotalAmount,
		RawPayload:       stringPtr(r.RawPayload),
		CanonicalData:    stringPtr(r.CanonicalData),
		Items:            items,
		CreatedAt:        toProtoTimestamp(r.CreatedAt),
		UpdatedAt:        toProtoTimestamp(r.UpdatedAt),
	}
}

// =======================
//  FROM PROTO (PB -> Domain)
// =======================

func fromProtoTimestamp(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
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
	return &domain.Account{
		Name:          req.GetName(),
		Bank:          req.GetBank(),
		Type:          fromProtoAccountType(req.GetType()),
		Alias:         req.Alias,
		AnchorBalance: req.GetAnchorBalance(),
		AnchorDate:    time.Now(), // matches REST: anchor set at creation
	}
}

func fromProtoCreateTransactionRequest(req *pb.CreateTransactionRequest) *domain.Transaction {
	return &domain.Transaction{
		EmailID:         req.EmailId,
		AccountID:       req.AccountId,
		TxDate:          fromProtoTimestamp(req.TxDate),
		TxAmount:        req.TxAmount,
		TxCurrency:      req.TxCurrency,
		TxDirection:     req.TxDirection,
		TxDesc:          req.TxDesc,
		CategoryID:      req.CategoryId,
		Merchant:        req.Merchant,
		UserNotes:       req.UserNotes,
		Suggestions:     pq.StringArray(req.Suggestions),
		ForeignAmount:   req.ForeignAmount,
		ForeignCurrency: req.ForeignCurrency,
		ExchangeRate:    req.ExchangeRate,
	}
}

func fromProtoReceiptProvider(p pb.ReceiptProvider) domain.ReceiptProvider {
	switch p {
	case pb.ReceiptProvider_RECEIPT_PROVIDER_GEMINI:
		return domain.ProviderGemini
	case pb.ReceiptProvider_RECEIPT_PROVIDER_LOCAL:
		return domain.ProviderLocal
	default:
		// Mirror REST default when provider not specified.
		return domain.ProviderGemini
	}
}

// =======================
//  HELPERS
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
