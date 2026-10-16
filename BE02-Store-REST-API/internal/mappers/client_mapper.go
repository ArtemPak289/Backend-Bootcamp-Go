package mappers

import (
	"errors"
	"store_api/internal/models/dao"
	"store_api/internal/models/dto"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ClientMapper handles conversions between Client DAO and DTO.
type ClientMapper struct {
	addrMapper *AddressMapper
}

func NewClientMapper() *ClientMapper {
	return &ClientMapper{
		addrMapper: NewAddressMapper(),
	}
}

func (m *ClientMapper) RequestDTOToDAO(req *dto.ClientRequestDTO, addressID uuid.UUID) (*dao.ClientDAO, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	var id uuid.UUID
	if req.ID != nil {
		id = *req.ID
	} else {
		id = uuid.New()
	}

	bday, err := parseDate(req.Birthday)
	if err != nil {
		return nil, errors.New("invalid birthday format, expected YYYY-MM-DD")
	}

	regDate := time.Now().UTC()
	if req.RegistrationDate != nil && strings.TrimSpace(*req.RegistrationDate) != "" {
		parsedReg, err := parseDateTime(*req.RegistrationDate)
		if err == nil {
			regDate = parsedReg
		}
	}

	return &dao.ClientDAO{
		ID:               id,
		ClientName:       strings.TrimSpace(req.ClientName),
		ClientSurname:    strings.TrimSpace(req.ClientSurname),
		Birthday:         bday,
		Gender:           strings.TrimSpace(req.Gender),
		RegistrationDate: regDate,
		AddressID:        addressID,
	}, nil
}

func (m *ClientMapper) DAOToDTO(d *dao.ClientDAO, addr *dao.AddressDAO) *dto.ClientResponseDTO {
	if d == nil {
		return nil
	}

	var addrDTO *dto.AddressDTO
	if addr != nil {
		addrDTO = m.addrMapper.DAOToDTO(addr)
	}

	return &dto.ClientResponseDTO{
		ID:               d.ID,
		ClientName:       d.ClientName,
		ClientSurname:    d.ClientSurname,
		Birthday:         d.Birthday.Format("2006-01-02"),
		Gender:           d.Gender,
		RegistrationDate: d.RegistrationDate.Format(time.RFC3339),
		AddressID:        d.AddressID,
		Address:          addrDTO,
	}
}

func (m *ClientMapper) DAOsToDTOs(daos []dao.ClientDAO, addrMap map[uuid.UUID]*dao.AddressDAO) []dto.ClientResponseDTO {
	if daos == nil {
		return []dto.ClientResponseDTO{}
	}
	result := make([]dto.ClientResponseDTO, len(daos))
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

func parseDate(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	formats := []string{
		"2006-01-02",
		time.RFC3339,
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("cannot parse date")
}

func parseDateTime(val string) (time.Time, error) {
	val = strings.TrimSpace(val)
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("cannot parse datetime")
}
