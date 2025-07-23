package handlers

import (
	"ariand/internal/service"
	"net/http"
)

type ReceiptHandler struct {
	service service.ReceiptService
}

// ParseReceipt godoc
// @Summary      Parse a receipt image
// @Description  Uploads a receipt image (JPEG or PNG) and returns the parsed, structured data.
// @Tags         receipts
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData  file   true   "Receipt image file"
// @Success      200  {object}  service.Receipt
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/receipts/parse [post]
// @Security     BearerAuth
func (h *ReceiptHandler) ParseReceipt(w http.ResponseWriter, r *http.Request) {
	// Max file size: 10 MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		NewHTTPError(http.StatusBadRequest, "file is too large").Write(w)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		NewHTTPError(http.StatusBadRequest, "invalid file upload").Write(w)
		return
	}
	defer file.Close()

	receipt, err := h.service.Parse(r.Context(), file, handler.Filename)
	if err != nil {
		NewHTTPError(http.StatusInternalServerError, err.Error()).Write(w)
		return
	}

	writeJSON(w, http.StatusOK, receipt)
}
