package mappers

import (
	"errors"
	"store_api/internal/models/dao"
	"store_api/internal/models/dto"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProductMapper handles conversions between Product DAO and DTO.
type ProductMapper struct{}

func NewProductMapper() *ProductMapper {
	return &ProductMapper{}
}

func (m *ProductMapper) RequestDTOToDAO(req *dto.ProductRequestDTO) (*dao.ProductDAO, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	var id uuid.UUID
	if req.ID != nil {
		id = *req.ID
	} else {
		id = uuid.New()
	}

	lastUpdate := time.Now().UTC()
	if req.LastUpdateDate != nil && strings.TrimSpace(*req.LastUpdateDate) != "" {
		if parsed, err := parseDateTime(*req.LastUpdateDate); err == nil {
			lastUpdate = parsed
		}
	}

	return &dao.ProductDAO{
		ID:             id,
		Name:           strings.TrimSpace(req.Name),
		Category:       strings.TrimSpace(req.Category),
		Price:          req.Price,
		AvailableStock: req.AvailableStock,
		LastUpdateDate: lastUpdate,
		SupplierID:     req.SupplierID,
		ImageID:        req.ImageID,
	}, nil
}

func (m *ProductMapper) DAOToDTO(d *dao.ProductDAO) *dto.ProductResponseDTO {
	if d == nil {
		return nil
	}

	return &dto.ProductResponseDTO{
		ID:             d.ID,
		Name:           d.Name,
		Category:       d.Category,
		Price:          d.Price,
		AvailableStock: d.AvailableStock,
		LastUpdateDate: d.LastUpdateDate.Format(time.RFC3339),
		SupplierID:     d.SupplierID,
		ImageID:        d.ImageID,
	}
}

func (m *ProductMapper) DAOsToDTOs(daos []dao.ProductDAO) []dto.ProductResponseDTO {
	if daos == nil {
		return []dto.ProductResponseDTO{}
	}
	result := make([]dto.ProductResponseDTO, len(daos))
	for i, d := range daos {
		dtoItem := m.DAOToDTO(&d)
		if dtoItem != nil {
			result[i] = *dtoItem
		}
	}
	return result
}
