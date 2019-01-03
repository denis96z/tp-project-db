package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"tp-project-db/models"
)

func (srv *Server) createForum(ctx *fasthttp.RequestCtx) {
	var forum models.Forum
	if err := srv.ReadBody(ctx, &forum); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	if err := srv.components.ForumValidator.Validate(&forum); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	if err := srv.components.ForumRepository.CreateForum(&forum); err != nil {
		if err.HttpStatus == http.StatusConflict {
			srv.WriteJSON(ctx, err.HttpStatus, &forum)
		} else {
			srv.WriteError(ctx, err)
		}
		return
	}

	srv.WriteJSON(ctx, http.StatusCreated, &forum)
}

func (srv *Server) findForumBySlug(ctx *fasthttp.RequestCtx) {
	slug := srv.readSlug(ctx)

	forum := models.Forum{
		Slug: slug,
	}
	if err := srv.components.ForumRepository.FindForumBySlug(&forum); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, &forum)
}
