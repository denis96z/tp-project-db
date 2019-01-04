package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"tp-project-db/models"
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
