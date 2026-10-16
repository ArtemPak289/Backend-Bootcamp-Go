package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"store_api/internal/mappers"
	"store_api/internal/models/dto"
	"store_api/internal/repository"
)

type ProductHandler struct {
	productRepo  repository.ProductRepository
	supplierRepo repository.SupplierRepository
	imageRepo    repository.ImageRepository
	mapper       *mappers.ProductMapper
}

func NewProductHandler(
	productRepo repository.ProductRepository,
	supplierRepo repository.SupplierRepository,
	imageRepo repository.ImageRepository,
	mapper *mappers.ProductMapper,
) *ProductHandler {
	return &ProductHandler{
		productRepo:  productRepo,
		supplierRepo: supplierRepo,
		imageRepo:    imageRepo,
		mapper:       mapper,
	}
}

// AddProduct godoc
// @Summary Add a new product
// @Description Creates a new appliance product
// @Tags Products
// @Accept json
// @Produce json
// @Param product body dto.ProductRequestDTO true "Product data"
// @Success 201 {object} dto.ProductResponseDTO
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/products [post]
func (h *ProductHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	var req dto.ProductRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if strings.TrimSpace(req.Category) == "" {
		respondError(w, http.StatusBadRequest, "category is required")
		return
	}
	if req.Price < 0 {
		respondError(w, http.StatusBadRequest, "price cannot be negative")
		return
	}
	if req.AvailableStock < 0 {
		respondError(w, http.StatusBadRequest, "available_stock cannot be negative")
		return
	}

	if req.SupplierID != nil {
		if _, err := h.supplierRepo.GetByID(r.Context(), *req.SupplierID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				respondError(w, http.StatusBadRequest, "supplier not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	if req.ImageID != nil {
		if _, err := h.imageRepo.GetByID(r.Context(), *req.ImageID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				respondError(w, http.StatusBadRequest, "image not found")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	daoItem, err := h.mapper.RequestDTOToDAO(&req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.productRepo.Create(r.Context(), daoItem); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create product: "+err.Error())
		return
	}

	res := h.mapper.DAOToDTO(daoItem)
	respondJSON(w, http.StatusCreated, res)
}

// DecreaseQuantity godoc
// @Summary Decrease product quantity
// @Description Decreases the stock count of a specific product by given amount
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Param input body dto.ProductDecreaseStockDTO false "Decrease payload"
// @Param amount query int false "Amount to decrease (if not in body)"
// @Success 200 {object} dto.ProductResponseDTO "Updated product"
// @Failure 400 {object} dto.ErrorResponse "Insufficient stock or invalid input"
// @Failure 404 {object} dto.ErrorResponse "Product not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/products/{id}/decrease [patch]
func (h *ProductHandler) DecreaseQuantity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid product id format")
		return
	}

	amount := 0
	if r.Body != nil && r.ContentLength > 0 {
		var req dto.ProductDecreaseStockDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			if req.Amount > 0 {
				amount = req.Amount
			} else if req.DecreaseBy > 0 {
				amount = req.DecreaseBy
			} else if req.Quantity > 0 {
				amount = req.Quantity
			}
		}
	}

	if amount <= 0 {
		if queryAmt := r.URL.Query().Get("amount"); queryAmt != "" {
			if a, err := strconv.Atoi(queryAmt); err == nil {
				amount = a
			}
		}
	}

	if amount <= 0 {
		respondError(w, http.StatusBadRequest, "amount to decrease must be a positive integer (> 0)")
		return
	}

	updated, err := h.productRepo.DecreaseStock(r.Context(), id, amount)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product not found")
			return
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			respondError(w, http.StatusBadRequest, "insufficient stock to decrease by requested amount")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := h.mapper.DAOToDTO(updated)
	respondJSON(w, http.StatusOK, res)
}

// GetProductByID godoc
// @Summary Get product by ID
// @Description Retrieves product details by ID
// @Tags Products
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} dto.ProductResponseDTO
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 404 {object} dto.ErrorResponse "Product not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid product id format")
		return
	}

	prod, err := h.productRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := h.mapper.DAOToDTO(prod)
	respondJSON(w, http.StatusOK, res)
}

// GetAvailableProducts godoc
// @Summary Get all available products
// @Description Retrieves all products that have available stock > 0
// @Tags Products
// @Produce json
// @Param all query bool false "Return all products including out-of-stock"
// @Success 200 {array} dto.ProductResponseDTO "List of products (empty if none)"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/products [get]
func (h *ProductHandler) GetAvailableProducts(w http.ResponseWriter, r *http.Request) {
	var prods []repository.ProductRepository
	_ = prods

	showAll := r.URL.Query().Get("all") == "true"
	var productsList []daoProduct
	_ = productsList

	if showAll {
		list, err := h.productRepo.GetAll(r.Context())
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		res := h.mapper.DAOsToDTOs(list)
		respondJSON(w, http.StatusOK, res)
		return
	}

	list, err := h.productRepo.GetAllAvailable(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	res := h.mapper.DAOsToDTOs(list)
	respondJSON(w, http.StatusOK, res)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Deletes a product by its unique identifier
// @Tags Products
// @Produce json
// @Param id path string true "Product ID (UUID)"
// @Success 200 {object} dto.MessageResponse "Product successfully deleted"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 404 {object} dto.ErrorResponse "Product not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid product id format")
		return
	}

	if err := h.productRepo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, dto.MessageResponse{Message: "product deleted successfully"})
}

type daoProduct = struct{}
