package services

import (
	"database/sql"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"tp-project-db/models"
	"tp-project-db/repositories"
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

func (srv *Server) findThreadsByForum(ctx *fasthttp.RequestCtx) {
	since := models.NullTimestamp{
		Valid: true,
	}
	value := ctx.QueryArgs().Peek("since")
	if len(value) == 0 {
		since.Valid = false
	}
	_ = since.Timestamp.UnmarshalText(ctx.QueryArgs().Peek("since"))

	limit, err := ctx.QueryArgs().GetUint("limit")
	if err != nil {
		limit = 0
	}

	args := repositories.ForumThreadsSearchArgs{
		Forum: ctx.UserValue("slug").(string),
		Since: since,
		Desc:  ctx.QueryArgs().GetBool("desc"),
		Limit: limit,
	}
	threads, searchErr := srv.components.ThreadRepository.FindThreadsByForum(&args)
	if searchErr != nil {
		srv.WriteError(ctx, searchErr.HttpStatus)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, threads)
}

func (srv *Server) updateThread(ctx *fasthttp.RequestCtx) {
	var threadUpdate models.ThreadUpdate
	srv.ReadBody(ctx, &threadUpdate)

	thread := models.Thread{
		Title:   threadUpdate.Title,
		Message: threadUpdate.Message,
	}

	slug := ctx.UserValue("slug_or_id").(string)
	id, err := strconv.ParseInt(slug, 10, 32)
	if err == nil {
		thread.ID = int32(id)
		if err := srv.components.ThreadRepository.UpdateThreadByID(&thread); err != nil {
			srv.WriteError(ctx, err.HttpStatus)
			return
		}
	} else {
		thread.Slug = models.NullString{
			Valid:  true,
			String: slug,
		}
		if err := srv.components.ThreadRepository.UpdateThreadBySlug(&thread); err != nil {
			srv.WriteError(ctx, err.HttpStatus)
			return
		}
	}

	srv.WriteJSON(ctx, http.StatusOK, &thread)
}
