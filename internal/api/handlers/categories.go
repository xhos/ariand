package handlers

import (
	"ariand/internal/db"
	"encoding/json"
	"errors"
	"net/http"
)

type CategoryHandler struct{ Store db.Store }

// List godoc
// @Summary      List all categories
// @Description  Returns a list of all transaction categories.
// @Tags         categories
// @Produce      json
// @Success      200  {array}   domain.Category
// @Failure      500  {object}  ErrorResponse
// @Router       /api/categories [get]
// @Security     BearerAuth
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Store.ListCategories(r.Context())
	if err != nil {
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

// Get godoc
// @Summary      Get a single category
// @Description  Retrieves a category by its numeric ID.
// @Tags         categories
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  domain.Category
// @Failure      400  {object}  ErrorResponse "invalid id format"
// @Failure      404  {object}  ErrorResponse "category not found"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/categories/{id} [get]
// @Security     BearerAuth
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

	category, err := h.Store.GetCategory(r.Context(), id)
	switch {
	case errors.Is(err, db.ErrNotFound):
		notFound(w)
	case err != nil:
		internalErr(w)
	default:
		writeJSON(w, http.StatusOK, category)
	}
}

type CreateCategoryRequest struct {
	Slug  string `json:"slug" example:"food.groceries"`
	Label string `json:"label" example:"Groceries"`
	Color string `json:"color,omitempty" example:"#FFD700"`
}

type CreateCategoryResponse struct {
	ID int64 `json:"id" example:"101"`
}

// Create godoc
// @Summary      Create a new category
// @Description  Adds a new category to the database. If color is omitted, a random one is assigned.
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        category body      CreateCategoryRequest   true  "Category object"
// @Success      201      {object}  CreateCategoryResponse
// @Failure      400      {object}  ErrorResponse "invalid request body"
// @Failure      500      {object}  ErrorResponse
// @Router       /api/categories [post]
// @Security     BearerAuth
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var in CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		badRequest(w, "invalid json")
		return
	}

	if in.Color == "" {
		in.Color = randomHexColor()
	}

	id, err := h.Store.CreateCategory(r.Context(), in.Slug, in.Label, in.Color)
	if err != nil {
		internalErr(w)
		return
	}

	writeJSON(w, http.StatusCreated, CreateCategoryResponse{ID: id})
}

type UpdateCategoryRequest map[string]any

// Patch godoc
// @Summary      Update a category
// @Description  Partially updates a category's fields. Only the provided fields will be changed.
// @Tags         categories
// @Accept       json
// @Param        id     path      int                    true  "Category ID"
// @Param        fields body      UpdateCategoryRequest  true  "Fields to update (e.g., {\"label\": \"New Label\"})"
// @Success      204
// @Failure      400  {object}  ErrorResponse "invalid request body"
// @Failure      404  {object}  ErrorResponse "category not found"
// @Failure      500  {object}  ErrorResponse
// @Router       /api/categories/{id} [patch]
// @Security     BearerAuth
func (h *CategoryHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		badRequest(w, "invalid json")
		return
	}

	if err := h.Store.UpdateCategory(r.Context(), id, fields); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			notFound(w)
			return
		}
		internalErr(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete godoc
// @Summary      Delete a category
// @Description  Deletes a category by its numeric ID.
// @Tags         categories
// @Param        id  path      int  true  "Category ID"
// @Success      204
// @Failure      400 {object}  ErrorResponse "invalid id format"
// @Failure      404 {object}  ErrorResponse "category not found"
// @Failure      500 {object}  ErrorResponse
// @Router       /api/categories/{id} [delete]
// @Security     BearerAuth
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		badRequest(w, "invalid id format")
		return
	}

	if err := h.Store.DeleteCategory(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			notFound(w)
			return
		}
		internalErr(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
