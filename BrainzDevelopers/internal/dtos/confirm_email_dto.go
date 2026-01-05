package dtos

type ConfirmEmailCode struct {
	Token string `json:"token,required" vd:"len($)=32"`
	Code  string `json:"code,required" vd:"len($)=6"`
}
