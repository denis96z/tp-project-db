package models

//go:generate easyjson

//easyjson:json
type Post struct {
	ID               int64         `json:"id"`
	ParentID         int64         `json:"parent"`
	Author           string        `json:"author"`
	Forum            string        `json:"forum"`
	Thread           int32         `json:"thread"`
	Message          string        `json:"message"`
	CreatedTimestamp NullTimestamp `json:"created"`
	IsEdited         bool          `json:"isEdited"`
}

//easyjson:json
type Posts []Post

//easyjson:json
type PostFull map[string]interface{}
