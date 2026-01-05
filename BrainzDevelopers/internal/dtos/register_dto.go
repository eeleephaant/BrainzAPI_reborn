package dtos

type RegisterDto struct {
	Email    string `json:"email,required" vd:"email($)"`
	Password string `json:"password,required" vd:"len($)>12 && len($)<256"`
}
