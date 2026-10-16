package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"store_api/internal/mappers"
	"store_api/internal/models/dao"
	"store_api/internal/models/dto"
	"store_api/internal/repository"

	"github.com/google/uuid"
)

type ClientHandler struct {
	clientRepo repository.ClientRepository
	addrRepo   repository.AddressRepository
	mapper     *mappers.ClientMapper
}

func NewClientHandler(
	clientRepo repository.ClientRepository,
	addrRepo repository.AddressRepository,
	mapper *mappers.ClientMapper,
) *ClientHandler {
	return &ClientHandler{
		clientRepo: clientRepo,
		addrRepo:   addrRepo,
		mapper:     mapper,
	}
}

// AddClient godoc
// @Summary Add a new client
// @Description Creates a new client in the database with address verification
// @Tags Clients
// @Accept json
// @Produce json
// @Param client body dto.ClientRequestDTO true "Client data"
// @Success 201 {object} dto.ClientResponseDTO
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/clients [post]
func (h *ClientHandler) AddClient(w http.ResponseWriter, r *http.Request) {
	var req dto.ClientRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if strings.TrimSpace(req.ClientName) == "" {
		respondError(w, http.StatusBadRequest, "client_name is required")
		return
	}
	if strings.TrimSpace(req.ClientSurname) == "" {
		respondError(w, http.StatusBadRequest, "client_surname is required")
		return
	}
	if strings.TrimSpace(req.Birthday) == "" {
		respondError(w, http.StatusBadRequest, "birthday is required (format YYYY-MM-DD)")
		return
	}
	if strings.TrimSpace(req.Gender) == "" {
		respondError(w, http.StatusBadRequest, "gender is required")
		return
	}

	var addressID uuid.UUID
	var addrDAO *dao.AddressDAO

	if req.AddressID != nil {
		existingAddr, err := h.addrRepo.GetByID(r.Context(), *req.AddressID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				respondError(w, http.StatusBadRequest, "referenced address_id does not exist")
				return
			}
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		addressID = existingAddr.ID
		addrDAO = existingAddr
	} else if req.Address != nil {
		country := strings.TrimSpace(req.Address.Country)
		city := strings.TrimSpace(req.Address.City)
		street := strings.TrimSpace(req.Address.Street)
		if country == "" || city == "" || street == "" {
			respondError(w, http.StatusBadRequest, "address country, city, and street must not be empty")
			return
		}
		newAddr, err := h.addrRepo.FindOrCreate(r.Context(), country, city, street)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save address: "+err.Error())
			return
		}
		addressID = newAddr.ID
		addrDAO = newAddr
	} else {
		respondError(w, http.StatusBadRequest, "address_id or address object is required")
		return
	}

	clientDAO, err := h.mapper.RequestDTOToDAO(&req, addressID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.clientRepo.Create(r.Context(), clientDAO); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create client: "+err.Error())
		return
	}

	res := h.mapper.DAOToDTO(clientDAO, addrDAO)
	respondJSON(w, http.StatusCreated, res)
}

// DeleteClient godoc
// @Summary Delete a client
// @Description Deletes a client by its unique identifier
// @Tags Clients
// @Produce json
// @Param id path string true "Client ID (UUID)"
// @Success 200 {object} dto.MessageResponse "Client successfully deleted"
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 404 {object} dto.ErrorResponse "Client not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/clients/{id} [delete]
func (h *ClientHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid client id format")
		return
	}

	if err := h.clientRepo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "client not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, dto.MessageResponse{Message: "client deleted successfully"})
}

// GetClientsByName godoc
// @Summary Get clients by first and last name
// @Description Searches clients by exact or case-insensitive first and last name
// @Tags Clients
// @Produce json
// @Param client_name query string true "First name"
// @Param client_surname query string true "Last name"
// @Success 200 {array} dto.ClientResponseDTO "List of clients (empty if none)"
// @Failure 400 {object} dto.ErrorResponse "Missing parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/clients/search [get]
func (h *ClientHandler) GetClientsByName(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("client_name"))
	surname := strings.TrimSpace(r.URL.Query().Get("client_surname"))

	if name == "" || surname == "" {
		respondError(w, http.StatusBadRequest, "client_name and client_surname query parameters are required")
		return
	}

	clients, err := h.clientRepo.GetByNameAndSurname(r.Context(), name, surname)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	addrMap := h.fetchAddressMap(r, clients)
	res := h.mapper.DAOsToDTOs(clients, addrMap)
	respondJSON(w, http.StatusOK, res)
}

// GetClients godoc
// @Summary Get all clients
// @Description Retrieves all clients with optional limit and offset pagination. Also filters by name if provided.
// @Tags Clients
// @Produce json
// @Param limit query int false "Maximum number of clients to return"
// @Param offset query int false "Number of clients to skip"
// @Param client_name query string false "First name filter"
// @Param client_surname query string false "Last name filter"
// @Success 200 {array} dto.ClientResponseDTO "List of clients (empty if none)"
// @Failure 400 {object} dto.ErrorResponse "Invalid pagination parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/clients [get]
func (h *ClientHandler) GetClients(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.URL.Query().Get("client_name"))
	surname := strings.TrimSpace(r.URL.Query().Get("client_surname"))

	// If both name and surname provided, behave as search
	if name != "" && surname != "" {
		h.GetClientsByName(w, r)
		return
	}

	var limitPtr *int
	var offsetPtr *int

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			respondError(w, http.StatusBadRequest, "limit must be a positive integer")
			return
		}
		limitPtr = &l
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			respondError(w, http.StatusBadRequest, "offset must be a non-negative integer")
			return
		}
		offsetPtr = &o
	}

	clients, err := h.clientRepo.GetAll(r.Context(), limitPtr, offsetPtr)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	addrMap := h.fetchAddressMap(r, clients)
	res := h.mapper.DAOsToDTOs(clients, addrMap)
	respondJSON(w, http.StatusOK, res)
}

// GetClientByID godoc
// @Summary Get client by ID
// @Description Retrieves a single client by its unique identifier
// @Tags Clients
// @Produce json
// @Param id path string true "Client ID (UUID)"
// @Success 200 {object} dto.ClientResponseDTO
// @Failure 400 {object} dto.ErrorResponse "Invalid UUID"
// @Failure 404 {object} dto.ErrorResponse "Client not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/clients/{id} [get]
func (h *ClientHandler) GetClientByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid client id format")
		return
	}

	client, err := h.clientRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "client not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	addr, _ := h.addrRepo.GetByID(r.Context(), client.AddressID)
	res := h.mapper.DAOToDTO(client, addr)
	respondJSON(w, http.StatusOK, res)
}

// ChangeClientAddress godoc
// @Summary Change client address
// @Description Updates the address of an existing client
// @Tags Clients
// @Accept json
// @Produce json
// @Param id path string true "Client ID (UUID)"
// @Param address body dto.ClientAddressUpdateDTO true "New address information"
// @Success 200 {object} dto.ClientResponseDTO "Updated client"
// @Failure 400 {object} dto.ErrorResponse "Validation error"
// @Failure 404 {object} dto.ErrorResponse "Client not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/clients/{id}/address [put]
func (h *ClientHandler) ChangeClientAddress(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := parseUUID(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid client id format")
		return
	}

	// Verify client exists
	existingClient, err := h.clientRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "client not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req dto.ClientAddressUpdateDTO
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
			respondError(w, http.StatusInternalServerError, "failed to update address: "+err.Error())
			return
		}
		newAddressID = addr.ID
		addrDAO = addr
	}

	updatedClient, err := h.clientRepo.UpdateAddress(r.Context(), existingClient.ID, newAddressID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update client address: "+err.Error())
		return
	}

	res := h.mapper.DAOToDTO(updatedClient, addrDAO)
	respondJSON(w, http.StatusOK, res)
}

func (h *ClientHandler) fetchAddressMap(r *http.Request, clients []dao.ClientDAO) map[uuid.UUID]*dao.AddressDAO {
	addrMap := make(map[uuid.UUID]*dao.AddressDAO)
	for _, c := range clients {
		if _, ok := addrMap[c.AddressID]; !ok {
			if addr, err := h.addrRepo.GetByID(r.Context(), c.AddressID); err == nil {
				addrMap[c.AddressID] = addr
			}
		}
	}
	return addrMap
}
