package mappers

import (
	"store_api/internal/models/dao"
	"store_api/internal/models/dto"
	"time"

	"github.com/google/uuid"
)

// AddressMapper handles conversions between Address DAO and DTO.
type AddressMapper struct{}

func NewAddressMapper() *AddressMapper {
	return &AddressMapper{}
}

func (m *AddressMapper) DAOToDTO(d *dao.AddressDAO) *dto.AddressDTO {
	if d == nil {
		return nil
	}
	id := d.ID
	return &dto.AddressDTO{
		ID:      &id,
		Country: d.Country,
		City:    d.City,
		Street:  d.Street,
	}
}

func (m *AddressMapper) DTOToDAO(d *dto.AddressDTO) *dao.AddressDAO {
	if d == nil {
		return nil
	}
	var id uuid.UUID
	if d.ID != nil {
		id = *d.ID
	} else {
		id = uuid.New()
	}
	return &dao.AddressDAO{
		ID:        id,
		Country:   d.Country,
		City:      d.City,
		Street:    d.Street,
		CreatedAt: time.Now().UTC(),
	}
}
