package models

import (
	"tp-project-db/consts"
	"tp-project-db/errs"
)

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

type PostValidator struct {
	err *errs.Error
}

func NewPostValidator() *PostValidator {
	return &PostValidator{
		err: errs.NewInvalidFormatError(ValidationErrMessage),
	}
}

func (v *PostValidator) Validate(post *Post) *errs.Error {
	if post.ParentID < 0 {
		return v.err
	}
	if post.Author == consts.EmptyString {
		return v.err
	}
	if post.Message == consts.EmptyString {
		return v.err
	}
	return nil
}

//easyjson:json
type Posts []Post

//easyjson:json
type PostFull map[string]interface{}

//easyjson:json
type PostUpdate struct {
	Message string `json:"message"`
}

type PostUpdateValidator struct {
	err *errs.Error
}

func NewPostUpdateValidator() *PostUpdateValidator {
	return &PostUpdateValidator{
		err: errs.NewInvalidFormatError(ValidationErrMessage),
	}
}

func (v *PostUpdateValidator) Validate(postUpdate *PostUpdate) *errs.Error {
	if postUpdate.Message == consts.EmptyString {
		return v.err
	}
	return nil
}
