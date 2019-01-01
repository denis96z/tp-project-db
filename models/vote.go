package models

//go:generate easyjson

//easyjson:json
type Vote struct {
	UserNickName string `json:"nickname"`
	VoiceValue   int32  `json:"voice"`
}

type VoteValidator struct {
	//TODO
}

func NewVoteValidator() *VoteValidator {
	return &VoteValidator{} //TODO
}
