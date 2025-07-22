package domain

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

type ReceiptProvider string

const (
	ProviderGemini ReceiptProvider = "gemini"
	ProviderLocal  ReceiptProvider = "local"
)

type ReceiptParseStatus string

const (
	StatusPending ReceiptParseStatus = "pending"
	StatusParsed  ReceiptParseStatus = "parsed"
	StatusFailed  ReceiptParseStatus = "failed"
)

type Receipt struct {
	ID             int64              `db:"id"               json:"id"`
	TransactionID  *int64             `db:"transaction_id"   json:"transactionId,omitempty"`
	Provider       ReceiptProvider    `db:"provider"         json:"provider"`
	ParseStatus    ReceiptParseStatus `db:"parse_status"     json:"parseStatus"`
	Merchant       *string            `db:"merchant"         json:"merchant,omitempty"`
	PurchaseDate   *time.Time         `db:"purchase_date"    json:"purchaseDate,omitempty"`
	TotalAmount    *float64           `db:"total_amount"     json:"totalAmount,omitempty"`
	Currency       *string            `db:"currency"         json:"currency,omitempty"`
	TaxAmount      *float64           `db:"tax_amount"       json:"taxAmount,omitempty"`
	RawPayload     types.JSONText     `db:"raw_payload"      json:"rawPayload,omitempty"`
	CanonicalData  types.JSONText     `db:"canonical_data"   json:"canonicalData,omitempty"`
	ImageURL       *string            `db:"image_url"        json:"imageUrl,omitempty"`
	ImageSHA256    []byte             `db:"image_sha256"     json:"-"`
	Lat            *float64           `db:"lat"              json:"lat,omitempty"`
	Lon            *float64           `db:"lon"              json:"lon,omitempty"`
	LocationSource *string            `db:"location_source"  json:"locationSource,omitempty"`
	LocationLabel  *string            `db:"location_label"   json:"locationLabel,omitempty"`
	CreatedAt      time.Time          `db:"created_at"       json:"createdAt"`
	UpdatedAt      time.Time          `db:"updated_at"       json:"updatedAt"`
	Items          []ReceiptItem      `json:"items,omitempty"`
}

type ReceiptItem struct {
	ID           int64     `db:"id"            json:"id"`
	ReceiptID    int64     `db:"receipt_id"    json:"-"`
	LineNo       *int      `db:"line_no"       json:"lineNo,omitempty"`
	Name         string    `db:"name"          json:"name"`
	Qty          *float64  `db:"qty"           json:"qty,omitempty"`
	UnitPrice    *float64  `db:"unit_price"    json:"unitPrice,omitempty"`
	LineTotal    *float64  `db:"line_total"    json:"lineTotal,omitempty"`
	SKU          *string   `db:"sku"           json:"sku,omitempty"`
	CategoryHint *string   `db:"category_hint" json:"categoryHint,omitempty"`
	CreatedAt    time.Time `db:"created_at"    json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at"    json:"updatedAt"`
}
