package dtos

type LoginDto struct {
	Email    string `json:"email,required" vd:"email($)"`
	Password string `json:"password,required" vd:"len($)>12 && len($)<256"`
}
