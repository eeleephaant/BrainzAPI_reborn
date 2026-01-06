package requests

type AddInstitutionRequest struct {
	Name     string `json:"institution_name" binding:"required"`
	SiteLink string `json:"instituion_site_link" binding:"required"`
}
