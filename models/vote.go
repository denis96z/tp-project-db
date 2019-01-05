package models

import (
	"tp-project-db/consts"
	"tp-project-db/errs"
)

//go:generate easyjson

//easyjson:json
type Vote struct {
	Author string `json:"nickname"`
	Thread int32  `json:"-"`
	Voice  int32  `json:"voice"`
}

type VoteValidator struct {
	err *errs.Error
}

func NewVoteValidator() *VoteValidator {
	return &VoteValidator{
		err: errs.NewInvalidFormatError(ValidationErrMessage),
	}
}

func (v *VoteValidator) Validate(vote *Vote) *errs.Error {
	if vote.Author == consts.EmptyString {
		return v.err
	}
	if vote.Voice != -1 && vote.Voice != 1 {
		return v.err
	}
	return nil
}
