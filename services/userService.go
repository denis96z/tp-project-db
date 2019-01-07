package services

import (
	"database/sql"
	"github.com/valyala/fasthttp"
	"net/http"
	"tp-project-db/models"
)

func (srv *Server) createUser(ctx *fasthttp.RequestCtx) {
	var user models.User
	srv.ReadBody(ctx, &user)

	user.Nickname = ctx.UserValue("nickname").(string)

	var existing sql.NullString
	status := srv.components.UserRepository.CreateUser(&user, &existing)
	if status != http.StatusCreated {
		ctx.SetStatusCode(status)
		ctx.Response.Header.SetContentType(JsonType)
		ctx.Response.SetBody([]byte(existing.String))
		return
	}

	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody(ctx.Request.Body())
	return
}
