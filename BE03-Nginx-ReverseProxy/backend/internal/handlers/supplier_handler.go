package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"store_api/internal/mappers"
	"store_api/internal/models/dao"
	"store_api/internal/models/dto"
	"store_api/internal/repository"

	"github.com/google/uuid"
)

type SupplierHandler struct {
	supplierRepo repository.SupplierRepository
	addrRepo     repository.AddressRepository
	mapper       *mappers.SupplierMapper
}

func NewSupplierHandler(
	supplierRepo repository.SupplierRepository,
	addrRepo repository.AddressRepository,
	mapper *mappers.SupplierMapper,
) *SupplierHandler {
	return &SupplierHandler{
		supplierRepo: supplierRepo,
		addrRepo:     addrRepo,
		mapper:       mapper,
	}
}

// AddSupplier godoc
// @Summary Add a new supplier
// @Description Creates a new appliance supplier with address
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param supplier body dto.SupplierRequestDTO true "Supplier data"
// @Success 201 {object} dto.SupplierResponseDTO
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/suppliers [post]
func (h *SupplierHandler) AddSupplier(w http.ResponseWriter, r *http.Request) {
	var req dto.SupplierRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if strings.TrimSpace(req.PhoneNumber) == "" {
		respondError(w, http.StatusBadRequest, "phone_number is required")
		return
	}

	var addressID uuid.UUID
	var addrDAO *dao.AddressDAO

	if req.AddressID != nil {
		addr, err := h.addrRepo.GetByID(r.Context(), *req.AddressID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				respondError(w, http.StatusBadRequest, "referenced address does not exist")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		addressID = addr.ID
		addrDAO = addr
	} else if req.Address != nil {
		country := strings.TrimSpace(req.Address.Country)
		city := strings.TrimSpace(req.Address.City)
		street := strings.TrimSpace(req.Address.Street)
		if country == "" || city == "" || street == "" {
			respondError(w, http.StatusBadRequest, "address country, city, and street must not be empty")
			return
		}
		addr, err := h.addrRepo.FindOrCreate(r.Context(), country, city, street)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save address: "+err.Error())
			return
		}
		addressID = addr.ID
		addrDAO = addr
	} else {
		respondError(w, http.StatusBadRequest, "address_id or address object is required")
		return
	}

	daoItem, err := h.mapper.RequestDTOToDAO(&req, addressID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.supplierRepo.Create(r.Context(), daoItem); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create supplier: "+err.Error())
		return
	}

	res := h.mapper.DAOToDTO(daoItem, addrDAO)
	respondJSON(w, http.StatusCreated, res)
}

// ChangeSupplierAddress godoc
// @Summary Change supplier address
// @Description Updates the address of an existing supplier
// @Tags Suppliers
// @Accept json
// @Produce json
// @Param id path string true "Supplier ID (UUID)"
// @Param address body dto.SupplierAddressUpdateDTO true "New address data"
// @Success 200 {object} dto.SupplierResponseDTO
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 404 {object} dto.ErrorResponse "Supplier not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/suppliers/{id}/address [put]
func (h *SupplierHandler) ChangeSupplierAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supplier id format")
		return
	}

	existing, err := h.supplierRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "supplier not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req dto.SupplierAddressUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	var newAddressID uuid.UUID
	var addrDAO *dao.AddressDAO

	if req.AddressID != nil {
		addr, err := h.addrRepo.GetByID(r.Context(), *req.AddressID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				respondError(w, http.StatusBadRequest, "referenced address does not exist")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		newAddressID = addr.ID
		addrDAO = addr
	} else {
		country := strings.TrimSpace(req.Country)
		city := strings.TrimSpace(req.City)
		street := strings.TrimSpace(req.Street)
		if country == "" || city == "" || street == "" {
			respondError(w, http.StatusBadRequest, "address country, city, and street must not be empty")
			return
		}
		addr, err := h.addrRepo.FindOrCreate(r.Context(), country, city, street)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save address: "+err.Error())
			return
		}
		newAddressID = addr.ID
		addrDAO = addr
	}

	updated, err := h.supplierRepo.UpdateAddress(r.Context(), existing.ID, newAddressID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update supplier address: "+err.Error())
		return
	}

	res := h.mapper.DAOToDTO(updated, addrDAO)
	respondJSON(w, http.StatusOK, res)
}

// DeleteSupplier godoc
// @Summary Delete a supplier
// @Description Deletes a supplier by its unique identifier
// @Tags Suppliers
// @Produce json
// @Param id path string true "Supplier ID (UUID)"
// @Success 200 {object} dto.MessageResponse "Supplier successfully deleted"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 404 {object} dto.ErrorResponse "Supplier not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/suppliers/{id} [delete]
func (h *SupplierHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supplier id format")
		return
	}

	if err := h.supplierRepo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "supplier not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, dto.MessageResponse{Message: "supplier deleted successfully"})
}

// GetAllSuppliers godoc
// @Summary Get all suppliers
// @Description Retrieves all registered suppliers
// @Tags Suppliers
// @Produce json
// @Success 200 {array} dto.SupplierResponseDTO "List of suppliers (empty if none)"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/suppliers [get]
func (h *SupplierHandler) GetAllSuppliers(w http.ResponseWriter, r *http.Request) {
	list, err := h.supplierRepo.GetAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	addrMap := h.fetchAddressMap(r, list)
	res := h.mapper.DAOsToDTOs(list, addrMap)
	respondJSON(w, http.StatusOK, res)
}

// GetSupplierByID godoc
// @Summary Get supplier by ID
// @Description Retrieves supplier information by ID
// @Tags Suppliers
// @Produce json
// @Param id path string true "Supplier ID (UUID)"
// @Success 200 {object} dto.SupplierResponseDTO
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 404 {object} dto.ErrorResponse "Supplier not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/suppliers/{id} [get]
func (h *SupplierHandler) GetSupplierByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid supplier id format")
		return
	}

	s, err := h.supplierRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "supplier not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	addr, _ := h.addrRepo.GetByID(r.Context(), s.AddressID)
	res := h.mapper.DAOToDTO(s, addr)
	respondJSON(w, http.StatusOK, res)
}

func (h *SupplierHandler) fetchAddressMap(r *http.Request, suppliers []dao.SupplierDAO) map[uuid.UUID]*dao.AddressDAO {
	addrMap := make(map[uuid.UUID]*dao.AddressDAO)
	for _, s := range suppliers {
		if _, ok := addrMap[s.AddressID]; !ok {
			if addr, err := h.addrRepo.GetByID(r.Context(), s.AddressID); err == nil {
				addrMap[s.AddressID] = addr
			}
		}
	}
	return addrMap
}
