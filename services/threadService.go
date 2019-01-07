package services

import (
	"database/sql"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"tp-project-db/models"
)

func (srv *Server) createThread(ctx *fasthttp.RequestCtx) {
	var thread models.Thread
	srv.ReadBody(ctx, &thread)

	thread.Forum = ctx.UserValue("slug").(string)

	var existing sql.NullString
	status := srv.components.ThreadRepository.CreateThread(&thread, &existing)

	if status == http.StatusCreated {
		srv.rwMtx.Lock()
		srv.status.NumThreads++
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

func (srv *Server) findThread(ctx *fasthttp.RequestCtx) {
	slugOrID := ctx.UserValue("slug_or_id").(string)
	id, err := strconv.ParseInt(slugOrID, 10, 32)

	var status int
	var existing string

	if err == nil {
		status = srv.components.ThreadRepository.FindThreadByID(int32(id), &existing)
	} else {
		status = srv.components.ThreadRepository.FindThreadBySlug(&slugOrID, &existing)
	}

	if status == http.StatusOK {
		ctx.SetStatusCode(status)
		ctx.Response.Header.SetContentType(JsonType)
		ctx.Response.SetBody([]byte(existing))
	} else {
		srv.WriteError(ctx, status)
	}
}
