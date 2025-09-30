package requests

type AddGroupRequest struct {
	Name          string `json:"group_name" binding:"required"`
	InstitutionId int    `json:"institution_id" binding:"required"`
}
