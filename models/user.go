package models

import (
	"regexp"
	"tp-project-db/consts"
	"tp-project-db/errs"
)

//go:generate easyjson

const (
	NickNamePattern = `^\w+$`
	EmailPattern    = `^.+@.+$`
)

//easyjson:json
type User struct {
	NickName string `json:"nickname"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

type UserValidator struct {
	nickNameRegexp *regexp.Regexp
	emailRegexp    *regexp.Regexp
	err            *errs.Error
}

func NewUserValidator() *UserValidator {
	return &UserValidator{
		nickNameRegexp: regexp.MustCompile(NickNamePattern),
		emailRegexp:    regexp.MustCompile(EmailPattern),
		err:            errs.NewInvalidFormatError(ValidationErrMessage),
	}
}

func (v *UserValidator) Validate(u *User) error {
	if !v.nickNameRegexp.MatchString(u.NickName) {
		return v.err
	}
	if u.FullName == consts.EmptyString {
		return v.err
	}
	if !v.emailRegexp.MatchString(u.Email) {
		return v.err
	}
	return nil
}

//easyjson:json
type UserUpdate struct {
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	About    string `json:"about"`
}

type UserUpdateValidator struct {
	emailRegexp *regexp.Regexp
	err         *errs.Error
}

func NewUserUpdateValidator() *UserUpdateValidator {
	return &UserUpdateValidator{
		emailRegexp: regexp.MustCompile(EmailPattern),
		err:         errs.NewInvalidFormatError(ValidationErrMessage),
	}
}

func (v *UserUpdateValidator) Validate(u *UserUpdate) error {
	if u.Email != consts.EmptyString && !v.emailRegexp.MatchString(u.Email) {
		return v.err
	}
	return nil
}

//easyjson:json
type Users []User
