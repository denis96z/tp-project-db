package models

import (
	"regexp"
	"tp-project-db/consts"
	"tp-project-db/errs"
)

//go:generate easyjson

const (
	ForumSlugPattern = `^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$`
)

//easyjson:json
type Forum struct {
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	AdminNickname string `json:"user"`
	NumThreads    int32  `json:"threads"`
	NumPosts      int64  `json:"posts"`
}

type ForumValidator struct {
	slugRegexp *regexp.Regexp
	err        *errs.Error
}

func NewForumValidator() *ForumValidator {
	return &ForumValidator{
		slugRegexp: regexp.MustCompile(ForumSlugPattern),
		err:        errs.NewInvalidFormatError(ValidationErrMessage),
	}
}

func (v *ForumValidator) Validate(forum *Forum) *errs.Error {
	if !v.slugRegexp.MatchString(forum.Slug) {
		return v.err
	}
	if forum.Title == consts.EmptyString {
		return v.err
	}
	if forum.AdminNickname == consts.EmptyString {
		return v.err
	}
	return nil
}
