package models

//go:generate easyjson

//easyjson:json
type Vote struct {
	User       string `json:"nickname"`
	ThreadID   int32  `json:"-"`
	ThreadSlug string `json:"-"`
	Voice      int32  `json:"voice"`
}
