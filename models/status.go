package models

//go:generate easyjson

//easyjson:json
type Status struct {
	NumUsers   int32 `json:"user"`
	NumForums  int32 `json:"forum"`
	NumThreads int32 `json:"thread"`
	NumPosts   int64 `json:"post"`
}
