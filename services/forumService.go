package services

import (
	"database/sql"
	"github.com/valyala/fasthttp"
	"net/http"
	"tp-project-db/models"
)

func (srv *Server) createForum(ctx *fasthttp.RequestCtx) {
	var forum models.Forum
	srv.ReadBody(ctx, &forum)

	var existing sql.NullString
	status := srv.components.ForumRepository.CreateForum(&forum, &existing)

	if status == http.StatusCreated {
		srv.rwMtx.Lock()
		srv.status.NumForums++
		srv.rwMtx.Unlock()
	}

	if existing.Valid {
		ctx.SetStatusCode(status)
		ctx.Response.Header.SetContentType(JsonType)
		ctx.Response.SetBody([]byte(existing.String))
	} else {
		srv.WriteError(ctx, status)
	}
}

func (srv *Server) findForum(ctx *fasthttp.RequestCtx) {
	forum := models.Forum{
		Slug: ctx.UserValue("slug").(string),
	}
	if err := srv.components.ForumRepository.FindForum(&forum); err != nil {
		srv.WriteError(ctx, err.HttpStatus)
		return
	}
	srv.WriteJSON(ctx, http.StatusOK, &forum)
}
