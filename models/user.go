package models

//go:generate easyjson

//easyjson:json
type User struct {
	Nickname string `json:"nickname"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

//easyjson:json
type UserUpdate struct {
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

//easyjson:json
type Users []User
