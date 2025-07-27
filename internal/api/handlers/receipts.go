package handlers

import (
	"ariand/internal/db"
	"ariand/internal/domain"
	"ariand/internal/service"
	"errors"
	"net/http"
)

type ReceiptHandler struct {
	service service.ReceiptService
}

// Upload handles the manual linking of a receipt to a specific transaction.
// @Summary      Upload a receipt for a transaction (Manual Link)
// @Description  Parses a receipt image, stores its data, and forcibly links it to an existing transaction.
// @Tags         transactions
// @Accept       multipart/form-data
// @Produce      json
// @Param        id        path      int    true   "Transaction ID"
// @Param        file      formData  file   true   "Receipt image file (e.g., JPEG, PNG)"
// @Param        provider  query     string false  "Parsing provider to use" Enums(gemini, local)
// @Success      201       {object}  domain.Receipt
// @Failure      400       {object}  ErrorResponse "Invalid transaction ID or missing/invalid file"
// @Failure      404       {object}  ErrorResponse "Transaction not found"
// @Failure      409       {object}  ErrorResponse "Transaction already has a receipt"
// @Failure      413       {object}  ErrorResponse "File is too large"
// @Failure      500       {object}  ErrorResponse "Internal server error or parser service failure"
// @Router       /api/transactions/{id}/receipt [post]
// @Security     BearerAuth
func (h *ReceiptHandler) Upload(r *http.Request) (any, *HTTPError) {
	transactionID, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid transaction id format")
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, NewHTTPError(http.StatusRequestEntityTooLarge, "file is too large (max 10MB)")
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "a multipart file with key 'file' is required")
	}
	defer file.Close()

	providerStr := r.URL.Query().Get("provider")
	if providerStr == "" {
		providerStr = string(domain.ProviderGemini)
	}
	provider := domain.ReceiptProvider(providerStr)
	if provider != domain.ProviderGemini && provider != domain.ProviderLocal {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid provider specified")
	}

	receipt, err := h.service.LinkManual(r.Context(), transactionID, file, handler.Filename, provider)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "transaction not found")
		}

		if errors.Is(err, db.ErrConflict) {
			return nil, NewHTTPError(http.StatusConflict, "transaction already has a receipt attached")
		}

		return nil, NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return receipt, nil
}

// Match handles the "smart matching" of a receipt to the best possible transaction.
// @Summary      Upload a receipt to find a matching transaction (Smart Match)
// @Description  Parses a receipt image and searches for the best transaction match based on date, amount, and merchant. Creates a receipt record with the match details.
// @Tags         receipts
// @Accept       multipart/form-data
// @Produce      json
// @Param        file      formData  file   true   "Receipt image file (e.g., JPEG, PNG)"
// @Param        provider  query     string false  "Parsing provider to use" Enums(gemini, local)
// @Success      201       {object}  domain.Receipt
// @Failure      400       {object}  ErrorResponse "Missing or invalid file"
// @Failure      413       {object}  ErrorResponse "File is too large"
// @Failure      500       {object}  ErrorResponse "Internal server error or parser service failure"
// @Router       /api/receipts/match [post]
// @Security     BearerAuth
func (h *ReceiptHandler) Match(r *http.Request) (any, *HTTPError) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, NewHTTPError(http.StatusRequestEntityTooLarge, "file is too large (max 10MB)")
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "a multipart file with key 'file' is required")
	}
	defer file.Close()

	providerStr := r.URL.Query().Get("provider")
	if providerStr == "" {
		providerStr = string(domain.ProviderGemini)
	}

	provider := domain.ReceiptProvider(providerStr)
	if provider != domain.ProviderGemini && provider != domain.ProviderLocal {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid provider specified")
	}

	receipt, err := h.service.MatchAndSuggest(r.Context(), file, handler.Filename, provider)
	if err != nil {
		return nil, NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return receipt, nil
}
