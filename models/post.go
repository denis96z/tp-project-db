package models

//go:generate easyjson

//easyjson:json
type Post struct {
	ID             int64  `json:"id"`
	AuthorNickname string `json:"author"`
	ForumSlug      string `json:"forum"`
	ThreadID       int32  `json:"thread"`
}
