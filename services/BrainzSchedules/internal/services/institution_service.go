package services

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/models"
	"brainz-api/internal/repositories"
	"context"
)

type InstitutionRepository interface {
	Create(ctx context.Context, institution *models.Institution) (int64, error)
	List(ctx context.Context) ([]*models.Institution, error)
}

type InstitutionService struct {
	ir InstitutionRepository
}

func NewInstitutionService(ir *repositories.InstitutionRepository) *InstitutionService {
	return &InstitutionService{ir}
}

var _ InstitutionRepository = (*repositories.InstitutionRepository)(nil)

func (is *InstitutionService) CreateNew(ctx context.Context, instDto dtos.InstitutionCreateDto) (int64, error) {
	institution := models.Institution{
		Name:     instDto.Name,
		SiteLink: &instDto.SiteLink,
	}
	createdID, err := is.ir.Create(ctx, &institution)
	if err != nil {
		return 0, err
	}
	return createdID, nil
}

func (is *InstitutionService) GetInstitutions(ctx context.Context) ([]dtos.InstitutionDto, error) {
	instList, err := is.ir.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]dtos.InstitutionDto, len(instList))

	for i, inst := range instList {
		site := ""
		if inst.SiteLink != nil {
			site = *inst.SiteLink
		}
		result[i] = dtos.InstitutionDto{
			Id:   inst.ID,
			Name: inst.Name,
			Site: site,
		}
	}

	return result, nil
}
