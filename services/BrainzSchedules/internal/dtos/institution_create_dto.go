package dtos

type InstitutionCreateDto struct {
	Name     string `json:"name,required" vd:"len($)<101"`
	SiteLink string `json:"site_link" vd:"len($)<256"`
}
