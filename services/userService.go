package services

import (
	"database/sql"
	"github.com/valyala/fasthttp"
	"net/http"
	"tp-project-db/models"
	"tp-project-db/repositories"
)

func (srv *Server) createUser(ctx *fasthttp.RequestCtx) {
	var user models.User
	srv.ReadBody(ctx, &user)

	user.Nickname = ctx.UserValue("nickname").(string)

	var existing string
	status := srv.components.UserRepository.CreateUser(&user, &existing)

	if status == http.StatusCreated {
		srv.rwMtx.Lock()
		srv.status.NumUsers++
		srv.rwMtx.Unlock()
	}

	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody([]byte(existing))
}

func (srv *Server) findUser(ctx *fasthttp.RequestCtx) {
	user := models.User{
		Nickname: ctx.UserValue("nickname").(string),
	}
	if err := srv.components.UserRepository.FindUser(&user); err != nil {
		srv.WriteError(ctx, err.HttpStatus)
		return
	}
	srv.WriteJSON(ctx, http.StatusOK, &user)
}

func (srv *Server) findUsersByForum(ctx *fasthttp.RequestCtx) {
	args := repositories.UsersByForumSearchArgs{
		Forum: ctx.UserValue("slug").(string),
		Since: string(ctx.QueryArgs().Peek("since")),
		Desc:  ctx.QueryArgs().GetBool("desc"),
		Limit: ctx.QueryArgs().GetUintOrZero("limit"),
	}
	users, err := srv.components.UserRepository.FindUsersByForum(&args)
	if err != nil {
		srv.WriteError(ctx, err.HttpStatus)
		return
	}
	srv.WriteJSON(ctx, http.StatusOK, users)
}

func (srv *Server) updateUser(ctx *fasthttp.RequestCtx) {
	var user models.User
	srv.ReadBody(ctx, &user)

	user.Nickname = ctx.UserValue("nickname").(string)

	var existing sql.NullString
	status := srv.components.UserRepository.UpdateUser(&user, &existing)

	if existing.Valid {
		ctx.SetStatusCode(status)
		ctx.Response.Header.SetContentType(JsonType)
		ctx.Response.SetBody([]byte(existing.String))
	} else {
		srv.WriteError(ctx, status)
	}
}
