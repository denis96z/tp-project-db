package models

//go:generate easyjson

//easyjson:json
type Forum struct {
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	AdminNickname string `json:"user"`
	NumThreads    int32  `json:"threads"`
	NumPosts      int64  `json:"posts"`
}

type ForumValidator struct {
	//TODO
}

func NewForumValidator() *ForumValidator {
	return &ForumValidator{} //TODO
}
