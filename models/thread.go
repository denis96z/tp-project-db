package models

//go:generate easyjson

//easyjson:json
type Thread struct {
	ID             int32  `json:"id"`
	Slug           string `json:"slug"`
	ForumSlug      string `json:"forum"`
	AuthorNickName string `json:"author"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Timestamp      string `json:"created"`
	NumVotes       int32  `json:"votes"`
}

type ThreadValidator struct {
	//TODO
}

func NewThreadValidator() *ThreadValidator {
	return &ThreadValidator{} //TODO
}

//easyjson:json
type ThreadUpdate struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ThreadUpdateValidator struct {
	//TODO
}

func NewThreadUpdateValidator() *ThreadUpdateValidator {
	return &ThreadUpdateValidator{} //TODO
}

//easyjson:json
type Threads []Thread
