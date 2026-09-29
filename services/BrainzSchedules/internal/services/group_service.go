package services

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/models"
	"brainz-api/internal/repositories"
	"context"
)

type GroupRepository interface {
	List(ctx context.Context) ([]models.Group, error)
	GetListForInstitution(ctx context.Context, instID int64) ([]models.Group, error)
}

type GroupService struct {
	gr GroupRepository
}

func NewGroupService(gr *repositories.GroupRepository) *GroupService {
	return &GroupService{gr}
}

var _ GroupRepository = (*repositories.GroupRepository)(nil)

func (gs *GroupService) GetGroups(ctx context.Context) ([]dtos.GroupDto, error) {
	groups, err := gs.gr.List(ctx)
	if err != nil {
		return nil, err
	}
	var groupDTOs []dtos.GroupDto
	for _, group := range groups {
		groupDTOs = append(groupDTOs, dtos.GroupDto{Id: group.ID, Name: group.Name})
	}
	return groupDTOs, nil
}

func (gs *GroupService) GetGroupsForInstitution(ctx context.Context, instID int64) ([]dtos.GroupDto, error) {
	groups, err := gs.gr.GetListForInstitution(ctx, instID)
	if err != nil {
		return nil, err
	}

	groupDTOs := make([]dtos.GroupDto, 0, len(groups))

	for _, g := range groups {
		groupDTOs = append(groupDTOs, dtos.GroupDto{Id: g.ID, Name: g.Name})
	}

	return groupDTOs, nil
}
