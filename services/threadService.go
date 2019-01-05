package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"tp-project-db/models"
	"tp-project-db/repositories"
)

func (srv *Server) createThread(ctx *fasthttp.RequestCtx) {
	var thread models.Thread
	if err := srv.ReadBody(ctx, &thread); err != nil {
		srv.WriteError(ctx, err)
		return
	}
	thread.Forum = srv.readSlug(ctx)

	if err := srv.components.ThreadValidator.Validate(&thread); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	if err := srv.components.ThreadRepository.CreateThread(&thread); err != nil {
		if err.HttpStatus == http.StatusConflict {
			srv.WriteJSON(ctx, err.HttpStatus, &thread)
		} else {
			srv.WriteError(ctx, err)
		}
		return
	}

	srv.WriteJSON(ctx, http.StatusCreated, &thread)
}

func (srv *Server) findThreadBySlugOrID(ctx *fasthttp.RequestCtx) {
	var thread models.Thread
	slug := srv.readSlugOrID(ctx)

	id, err := strconv.ParseInt(slug, 10, 32)
	if err == nil {
		thread.ID = int32(id)
		if err := srv.components.ThreadRepository.FindThreadByID(&thread); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	} else {
		thread.Slug = models.NullString{
			Valid:  true,
			String: slug,
		}
		if err := srv.components.ThreadRepository.FindThreadBySlug(&thread); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	}

	srv.WriteJSON(ctx, http.StatusOK, &thread)
}

func (srv *Server) findThreadsByForum(ctx *fasthttp.RequestCtx) {
	since := models.NullTimestamp{
		Valid: true,
	}
	err := srv.readSince(ctx, &since.Timestamp)
	if err != nil {
		since.Valid = false
	}

	args := repositories.ForumThreadsSearchArgs{
		Forum: srv.readSlug(ctx),
		Since: since,
		Desc:  srv.readDescFlag(ctx),
		Limit: srv.readLimit(ctx),
	}
	threads, searchErr := srv.components.ThreadRepository.FindThreadsByForum(&args)
	if searchErr != nil {
		srv.WriteError(ctx, searchErr)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, threads)
}

func (srv *Server) updateThread(ctx *fasthttp.RequestCtx) {
	var threadUpdate models.ThreadUpdate
	if err := srv.ReadBody(ctx, &threadUpdate); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	if err := srv.components.ThreadUpdateValidator.Validate(&threadUpdate); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	thread := models.Thread{
		Title:   threadUpdate.Title,
		Message: threadUpdate.Message,
	}

	slug := srv.readSlugOrID(ctx)
	id, err := strconv.ParseInt(slug, 10, 32)
	if err == nil {
		thread.ID = int32(id)
		if err := srv.components.ThreadRepository.UpdateThreadByID(&thread); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	} else {
		thread.Slug = models.NullString{
			Valid:  true,
			String: slug,
		}
		if err := srv.components.ThreadRepository.UpdateThreadBySlug(&thread); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	}

	srv.WriteJSON(ctx, http.StatusOK, &thread)
}
