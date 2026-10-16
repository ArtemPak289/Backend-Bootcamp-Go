package mappers

import (
	"store_api/internal/models/dao"
	"store_api/internal/models/dto"
)

// ImageMapper handles conversions between Image DAO and DTO.
type ImageMapper struct{}

func NewImageMapper() *ImageMapper {
	return &ImageMapper{}
}

func (m *ImageMapper) DAOToDTO(d *dao.ImageDAO) *dto.ImageResponseDTO {
	if d == nil {
		return nil
	}

	return &dto.ImageResponseDTO{
		ID:        d.ID,
		ProductID: d.ProductID,
		Size:      len(d.Image),
	}
}
