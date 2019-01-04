package models

import (
	"regexp"
	"tp-project-db/consts"
	"tp-project-db/errs"
)

//go:generate easyjson

const (
	ThreadSlugPattern = `^(\d|\w|-|_)*(\w|-|_)(\d|\w|-|_)*$`
)

//easyjson:json
type Thread struct {
	ID               int32         `json:"id"`
	Slug             string        `json:"slug"`
	Forum            string        `json:"forum"`
	Author           string        `json:"author"`
	Title            string        `json:"title"`
	Message          string        `json:"message"`
	CreatedTimestamp NullTimestamp `json:"created"`
	NumVotes         int32         `json:"votes"`
}

type ThreadValidator struct {
	slugRegexp *regexp.Regexp
	err        *errs.Error
}

func NewThreadValidator() *ThreadValidator {
	return &ThreadValidator{
		slugRegexp: regexp.MustCompile(ThreadSlugPattern),
		err:        errs.NewInvalidFormatError(ValidationErrMessage),
	}
}

func (v *ThreadValidator) Validate(thread *Thread) *errs.Error {
	if thread.Slug != consts.EmptyString && !v.slugRegexp.MatchString(thread.Slug) {
		return v.err
	}
	if thread.Title == consts.EmptyString {
		return v.err
	}
	if thread.Author == consts.EmptyString {
		return v.err
	}
	if thread.Message == consts.EmptyString {
		return v.err
	}
	return nil
}

//easyjson:json
type ThreadUpdate struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type ThreadUpdateValidator struct {
	err *errs.Error
}

func NewThreadUpdateValidator() *ThreadUpdateValidator {
	return &ThreadUpdateValidator{
		err: errs.NewInvalidFormatError(ValidationErrMessage),
	}
}

func (v *ThreadUpdateValidator) Validate(threadUpdate *ThreadUpdate) *errs.Error {
	return nil
}

//easyjson:json
type Threads []Thread
