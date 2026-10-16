package mappers

import (
	"errors"
	"store_api/internal/models/dao"
	"store_api/internal/models/dto"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SupplierMapper handles conversions between Supplier DAO and DTO.
type SupplierMapper struct {
	addrMapper *AddressMapper
}

func NewSupplierMapper() *SupplierMapper {
	return &SupplierMapper{
		addrMapper: NewAddressMapper(),
	}
}

func (m *SupplierMapper) RequestDTOToDAO(req *dto.SupplierRequestDTO, addressID uuid.UUID) (*dao.SupplierDAO, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	var id uuid.UUID
	if req.ID != nil {
		id = *req.ID
	} else {
		id = uuid.New()
	}

	return &dao.SupplierDAO{
		ID:          id,
		Name:        strings.TrimSpace(req.Name),
		AddressID:   addressID,
		PhoneNumber: strings.TrimSpace(req.PhoneNumber),
		CreatedAt:   time.Now().UTC(),
	}, nil
}

func (m *SupplierMapper) DAOToDTO(d *dao.SupplierDAO, addr *dao.AddressDAO) *dto.SupplierResponseDTO {
	if d == nil {
		return nil
	}

	var addrDTO *dto.AddressDTO
	if addr != nil {
		addrDTO = m.addrMapper.DAOToDTO(addr)
	}

	return &dto.SupplierResponseDTO{
		ID:          d.ID,
		Name:        d.Name,
		AddressID:   d.AddressID,
		Address:     addrDTO,
		PhoneNumber: d.PhoneNumber,
	}
}

func (m *SupplierMapper) DAOsToDTOs(daos []dao.SupplierDAO, addrMap map[uuid.UUID]*dao.AddressDAO) []dto.SupplierResponseDTO {
	if daos == nil {
		return []dto.SupplierResponseDTO{}
	}
	result := make([]dto.SupplierResponseDTO, len(daos))
	for i, d := range daos {
		var addr *dao.AddressDAO
		if addrMap != nil {
			addr = addrMap[d.AddressID]
		}
		dtoItem := m.DAOToDTO(&d, addr)
		if dtoItem != nil {
			result[i] = *dtoItem
		}
	}
	return result
}
