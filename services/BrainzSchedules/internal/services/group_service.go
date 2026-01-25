package services

import (
	"brainz-api/internal/dtos"
	"brainz-api/internal/repositories"
	"context"
)

type GroupService struct {
	gr *repositories.GroupRepository
}

func NewGroupService(gr *repositories.GroupRepository) *GroupService {
	return &GroupService{gr}
}

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

	for i, g := range groups {
		groupDTOs[i] = dtos.GroupDto{Id: g.ID, Name: g.Name}
	}

	return groupDTOs, nil
}
