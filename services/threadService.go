package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
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
		srv.WriteError(ctx, err)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, &thread)
}
