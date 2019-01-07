package services

import (
	"database/sql"
	"github.com/valyala/fasthttp"
	"tp-project-db/models"
)

func (srv *Server) createUser(ctx *fasthttp.RequestCtx) {
	var user models.User
	srv.ReadBody(ctx, &user)

	user.Nickname = ctx.UserValue("nickname").(string)

	var existing sql.NullString
	status := srv.components.UserRepository.CreateUser(&user, &existing)

	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody([]byte(existing.String))
	return
}
