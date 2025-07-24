// internal/db/postgres/queries/helpers.go
package queries

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"ariand/internal/domain"
)

// pull out all `db:"..."` tag names from a struct
func getDBFieldsFromStruct(t reflect.Type) []string {
	var cols []string
	for i := 0; i < t.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("db"); tag != "" && tag != "-" {
			cols = append(cols, tag)
		}
	}
	return cols
}

// build a comma‑separated list of "prefix.col" for each db tag
func buildSelectFields(t reflect.Type, prefix string) string {
	var out []string
	for _, col := range getDBFieldsFromStruct(t) {
		if prefix != "" {
			out = append(out, fmt.Sprintf("%s.%s", prefix, col))
		} else {
			out = append(out, col)
		}
	}
	return strings.Join(out, ", ")
}

// only allow updates to these struct tags
func getAllowedCols(t reflect.Type, excluded ...string) map[string]bool {
	ex := make(map[string]bool, len(excluded))
	for _, e := range excluded {
		ex[e] = true
	}
	allow := make(map[string]bool)
	for _, col := range getDBFieldsFromStruct(t) {
		if !ex[col] {
			allow[col] = true
		}
	}
	return allow
}

// convenience for the other tables
func getAccountFields() string  { return buildSelectFields(reflect.TypeOf(domain.Account{}), "") }
func getCategoryFields() string { return buildSelectFields(reflect.TypeOf(domain.Category{}), "") }
func getReceiptFields() string  { return buildSelectFields(reflect.TypeOf(domain.Receipt{}), "") }

// === the one with the join ===
func getTransactionFields() string {
	// grab only the real "transactions." columns
	all := getDBFieldsFromStruct(reflect.TypeOf(domain.Transaction{}))
	var cols []string
	for _, c := range all {
		// Skip the JOIN fields that don't exist in the transactions table
		if c == "category_slug" || c == "category_label" || c == "category_color" {
			continue
		}
		cols = append(cols, fmt.Sprintf("t.%s", c))
	}

	// now explicitly join the three category columns
	cols = append(cols,
		"c.slug  AS category_slug",
		"c.label AS category_label",
		"c.color AS category_color",
	)

	return strings.Join(cols, ", ")
}

// buildUpdateClauses — identical to before, but driven off reflect tags
func buildUpdateClauses(fields map[string]any, allowed map[string]bool) (string, []any, error) {
	if len(fields) == 0 {
		return "", nil, errors.New("no fields supplied")
	}
	var (
		parts []string
		args  []any
		i     = 1
	)
	for col, val := range fields {
		if !allowed[col] {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}
	if len(parts) == 0 {
		return "", nil, errors.New("no permitted columns in update")
	}
	return strings.Join(parts, ", "), args, nil
}

// allowed columns for each table
var (
	allowedTransactionCols = getAllowedCols(
		reflect.TypeOf(domain.Transaction{}),
		"id", "created_at", "updated_at", "balance_after",
		"category_slug", "category_label", "category_color", // exclude JOIN fields
	)
	allowedCategoryCols = getAllowedCols(
		reflect.TypeOf(domain.Category{}),
		"id", "created_at", "updated_at",
	)
	allowedReceiptCols = getAllowedCols(
		reflect.TypeOf(domain.Receipt{}),
		"id", "created_at", "updated_at", "image_sha256",
	)
	allowedAccountCols = getAllowedCols(
		reflect.TypeOf(domain.Account{}),
		"id", "created_at", "updated_at",
	)
)
