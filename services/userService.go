package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"tp-project-db/models"
	"tp-project-db/repositories"
)

func (srv *Server) createUser(ctx *fasthttp.RequestCtx) {
	var user models.User
	if err := srv.ReadBody(ctx, &user); err != nil {
		srv.WriteError(ctx, err)
		return
	}
	user.Nickname = srv.readNickname(ctx)

	if err := srv.components.UserValidator.Validate(&user); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	var existing models.Users
	if err := srv.components.UserRepository.CreateUser(&user, &existing); err != nil {
		if err.HttpStatus == http.StatusConflict {
			srv.WriteJSON(ctx, err.HttpStatus, &existing)
		} else {
			srv.WriteJSON(ctx, err.HttpStatus, &user)
		}
		return
	}

	srv.WriteJSON(ctx, http.StatusCreated, &user)
}

func (srv *Server) findUserByNickname(ctx *fasthttp.RequestCtx) {
	user := models.User{
		Nickname: srv.readNickname(ctx),
	}
	if err := srv.components.UserRepository.FindUserByNickname(&user); err != nil {
		srv.WriteError(ctx, err)
		return
	}
	srv.WriteJSON(ctx, http.StatusOK, &user)
}

func (srv *Server) findUsersByForum(ctx *fasthttp.RequestCtx) {
	args := repositories.UsersByForumSearchArgs{
		Forum: srv.readSlug(ctx),
		Since: srv.readSinceNickname(ctx),
		Desc:  srv.readDescFlag(ctx),
		Limit: srv.readLimit(ctx),
	}
	users, err := srv.components.UserRepository.FindUsersByForum(&args)
	if err != nil {
		srv.WriteError(ctx, err)
		return
	}
	srv.WriteJSON(ctx, http.StatusOK, users)
}

func (srv *Server) updateUserByNickname(ctx *fasthttp.RequestCtx) {
	var up models.UserUpdate
	if err := srv.ReadBody(ctx, &up); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	if err := srv.components.UserUpdateValidator.Validate(&up); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	user := models.User{
		Nickname: srv.readNickname(ctx),
		FullName: up.FullName,
		Email:    up.Email,
		About:    up.About,
	}
	if err := srv.components.UserRepository.UpdateUserByNickname(&user); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, &user)
}

func (srv *Server) deleteAllUsers(ctx *fasthttp.RequestCtx) {
	_ = srv.components.UserRepository.DeleteAllUsers()
}
