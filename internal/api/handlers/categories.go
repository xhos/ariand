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
func (h *CategoryHandler) List(r *http.Request) (any, *HTTPError) {
	categories, err := h.Store.ListCategories(r.Context())
	if err != nil {
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
	return categories, nil
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
func (h *CategoryHandler) Get(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	category, err := h.Store.GetCategory(r.Context(), id)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return category, nil
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
func (h *CategoryHandler) Create(r *http.Request) (any, *HTTPError) {
	var in CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid json")
	}

	if in.Color == "" {
		in.Color = randomHexColor()
	}

	id, err := h.Store.CreateCategory(r.Context(), in.Slug, in.Label, in.Color)
	if err != nil {
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return CreateCategoryResponse{ID: id}, nil
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
func (h *CategoryHandler) Patch(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	var fields map[string]any
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid json")
	}

	if err := h.Store.UpdateCategory(r.Context(), id, fields); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil, nil
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
func (h *CategoryHandler) Delete(r *http.Request) (any, *HTTPError) {
	id, err := parseIDFromRequest(r)
	if err != nil {
		return nil, NewHTTPError(http.StatusBadRequest, "invalid id format")
	}

	if err := h.Store.DeleteCategory(r.Context(), id); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, NewHTTPError(http.StatusNotFound, "resource not found")
		}
		return nil, NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return nil, nil
}
