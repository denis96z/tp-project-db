package models

//go:generate easyjson

//easyjson:json
type User struct {
	NickName string `json:"nickname"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

type UserValidator struct {
	//TODO
}

func NewUserValidator() *UserValidator {
	return &UserValidator{} //TODO
}

//easyjson:json
type UserUpdate struct {
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

type UserUpdateValidator struct {
	//TODO
}

func NewUserUpdateValidator() *UserUpdateValidator {
	return &UserUpdateValidator{} //TODO
}

//easyjson:json
type Users []User
