package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"store_api/internal/mappers"
	"store_api/internal/models/dao"
	"store_api/internal/models/dto"
	"store_api/internal/repository"

	"github.com/google/uuid"
)

type ImageHandler struct {
	imageRepo   repository.ImageRepository
	productRepo repository.ProductRepository
	mapper      *mappers.ImageMapper
}

func NewImageHandler(
	imageRepo repository.ImageRepository,
	productRepo repository.ProductRepository,
	mapper *mappers.ImageMapper,
) *ImageHandler {
	return &ImageHandler{
		imageRepo:   imageRepo,
		productRepo: productRepo,
		mapper:      mapper,
	}
}

// AddImage godoc
// @Summary Add an image
// @Description Uploads an image for a specific product. Supports multipart/form-data, JSON with base64, or raw octet-stream bytes.
// @Tags Images
// @Accept mpfd,json,octet-stream
// @Produce json
// @Param product_id query string false "Product ID (UUID) if sending raw bytes"
// @Param image formData file false "Image binary file (if multipart)"
// @Param json body dto.ImageUploadJSONDTO false "Image in base64 (if JSON)"
// @Success 201 {object} dto.ImageResponseDTO "Image uploaded successfully"
// @Failure 400 {object} dto.ErrorResponse "Validation error or invalid product ID"
// @Failure 404 {object} dto.ErrorResponse "Product not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/images [post]
func (h *ImageHandler) AddImage(w http.ResponseWriter, r *http.Request) {
	var productID uuid.UUID
	var imageBytes []byte

	// Check if product_id is in URL path (for /api/v1/products/{id}/images)
	if pathID, ok := mux.Vars(r)["id"]; ok {
		if id, err := parseUUID(pathID); err == nil {
			productID = id
		}
	}

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Parse multipart
		if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
			respondError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
			return
		}

		if productID == uuid.Nil {
			prodIDStr := r.FormValue("product_id")
			if prodIDStr == "" {
				prodIDStr = r.URL.Query().Get("product_id")
			}
			id, err := parseUUID(prodIDStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "valid product_id is required")
				return
			}
			productID = id
		}

		file, _, err := r.FormFile("image")
		if err != nil {
			file, _, err = r.FormFile("file")
		}
		if err != nil {
			respondError(w, http.StatusBadRequest, "image file field is required in form-data")
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			respondError(w, http.StatusBadRequest, "failed to read uploaded file: "+err.Error())
			return
		}
		imageBytes = data

	} else if strings.HasPrefix(contentType, "application/json") {
		var req dto.ImageUploadJSONDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid json payload: "+err.Error())
			return
		}

		if productID == uuid.Nil {
			if req.ProductID == uuid.Nil {
				respondError(w, http.StatusBadRequest, "product_id is required")
				return
			}
			productID = req.ProductID
		}

		decoded, err := base64.StdEncoding.DecodeString(req.Image)
		if err != nil {
			imageBytes = []byte(req.Image)
		} else {
			imageBytes = decoded
		}

	} else {
		// Binary octet-stream or raw
		if productID == uuid.Nil {
			prodIDStr := r.URL.Query().Get("product_id")
			if prodIDStr == "" {
				prodIDStr = r.Header.Get("X-Product-Id")
			}
			id, err := parseUUID(prodIDStr)
			if err != nil {
				respondError(w, http.StatusBadRequest, "product_id query parameter or header is required")
				return
			}
			productID = id
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			respondError(w, http.StatusBadRequest, "failed to read image body: "+err.Error())
			return
		}
		imageBytes = data
	}

	if len(imageBytes) == 0 {
		respondError(w, http.StatusBadRequest, "image data cannot be empty")
		return
	}

	// Verify product exists
	_, err := h.productRepo.GetByID(r.Context(), productID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	imgID := uuid.New()
	imgDAO := &dao.ImageDAO{
		ID:        imgID,
		Image:     imageBytes,
		ProductID: &productID,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.imageRepo.Create(r.Context(), imgDAO); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save image: "+err.Error())
		return
	}

	// Update product's image_id
	_ = h.productRepo.UpdateImageID(r.Context(), productID, &imgID)

	res := h.mapper.DAOToDTO(imgDAO)
	respondJSON(w, http.StatusCreated, res)
}

// ReplaceImage godoc
// @Summary Replace image
// @Description Replaces the byte array of an existing image by image ID
// @Tags Images
// @Accept json,plain,octet-stream
// @Produce json
// @Param id path string true "Image ID (UUID)"
// @Param body body dto.ImageReplaceDTO false "New replacement data/string"
// @Success 200 {object} dto.ImageResponseDTO "Image replaced successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID or empty replacement data"
// @Failure 404 {object} dto.ErrorResponse "Image not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/images/{id} [put]
func (h *ImageHandler) ReplaceImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid image id format")
		return
	}

	// Verify image exists
	if _, err := h.imageRepo.GetByID(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "image not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var newBytes []byte
	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "application/json") {
		var req dto.ImageReplaceDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			str := req.Image
			if str == "" {
				str = req.Data
			}
			if str != "" {
				if dec, err := base64.StdEncoding.DecodeString(str); err == nil {
					newBytes = dec
				} else {
					newBytes = []byte(str)
				}
			}
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err == nil {
			if file, _, err := r.FormFile("image"); err == nil {
				defer file.Close()
				newBytes, _ = io.ReadAll(file)
			} else if file, _, err := r.FormFile("file"); err == nil {
				defer file.Close()
				newBytes, _ = io.ReadAll(file)
			}
		}
	}

	if len(newBytes) == 0 {
		data, err := io.ReadAll(r.Body)
		if err == nil && len(data) > 0 {
			newBytes = data
		}
	}

	if len(newBytes) == 0 {
		respondError(w, http.StatusBadRequest, "replacement image content cannot be empty")
		return
	}

	updated, err := h.imageRepo.Update(r.Context(), id, newBytes)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "image not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	res := h.mapper.DAOToDTO(updated)
	respondJSON(w, http.StatusOK, res)
}

// DeleteImage godoc
// @Summary Delete an image
// @Description Deletes an image by image ID
// @Tags Images
// @Produce json
// @Param id path string true "Image ID (UUID)"
// @Success 200 {object} dto.MessageResponse "Image deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 404 {object} dto.ErrorResponse "Image not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/images/{id} [delete]
func (h *ImageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid image id format")
		return
	}

	if err := h.imageRepo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "image not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, dto.MessageResponse{Message: "image deleted successfully"})
}

// GetImageOfProduct godoc
// @Summary Get image of a specific product
// @Description Returns the byte array of an image for a specific product ID. The file will be automatically downloaded with application/octet-stream.
// @Tags Images
// @Produce octet-stream
// @Param id path string true "Product ID (UUID)"
// @Success 200 {file} binary "Image byte array"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 404 {object} dto.ErrorResponse "Product or image not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/products/{id}/image [get]
func (h *ImageHandler) GetImageOfProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	prodIDStr := vars["id"]
	if prodIDStr == "" {
		prodIDStr = vars["product_id"]
	}
	if prodIDStr == "" {
		prodIDStr = r.URL.Query().Get("product_id")
	}

	id, err := parseUUID(prodIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid product id format")
		return
	}

	// Verify product exists
	_, err = h.productRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	img, err := h.imageRepo.GetByProductID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "image not found for this product")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"product-image-%s.bin\"", id))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(img.Image)
}

// GetImageByID godoc
// @Summary Get image by image ID
// @Description Returns the byte array of an image by image ID. The file will be automatically downloaded with application/octet-stream.
// @Tags Images
// @Produce octet-stream
// @Param id path string true "Image ID (UUID)"
// @Success 200 {file} binary "Image byte array"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 404 {object} dto.ErrorResponse "Image not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/images/{id} [get]
func (h *ImageHandler) GetImageByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid image id format")
		return
	}

	img, err := h.imageRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "image not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"image-%s.bin\"", id))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(img.Image)
}
