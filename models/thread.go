package models

//go:generate easyjson

//easyjson:json
type Thread struct {
	ID               int32         `json:"id"`
	Slug             NullString    `json:"slug"`
	Forum            string        `json:"forum"`
	Author           string        `json:"author"`
	Title            string        `json:"title"`
	Message          string        `json:"message"`
	CreatedTimestamp NullTimestamp `json:"created"`
	NumVotes         int32         `json:"votes"`
}

//easyjson:json
type ThreadUpdate struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

//easyjson:json
type Threads []Thread
