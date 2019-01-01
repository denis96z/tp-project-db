package models

//go:generate easyjson

//easyjson:json
type Error struct {
	Message string `json:"message"`
}
