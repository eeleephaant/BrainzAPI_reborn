package dtos

type CreateKeyDto struct {
	Name string `json:"name,required" vd:"len($)>32 && len($)<256"`
}
