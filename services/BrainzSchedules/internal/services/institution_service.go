package services

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/models"
	"brainz-api/internal/repositories"
	"context"
)

type InstitutionService struct {
	ir *repositories.InstitutionRepository
}

func NewInstitutionService(ir *repositories.InstitutionRepository) *InstitutionService {
	return &InstitutionService{ir}
}

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
		result[i] = dtos.InstitutionDto{
			Id:   inst.ID,
			Name: inst.Name,
			Site: *inst.SiteLink,
		}
	}

	return result, nil
}
