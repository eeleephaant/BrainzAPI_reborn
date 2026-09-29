package dtos

import "brainz/common/permissions"

type RequestedPermissionDTO struct {
	Permission permissions.Permission `json:"perm,required"`
}
