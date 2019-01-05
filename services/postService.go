package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"tp-project-db/models"
)

func (srv *Server) createPost(ctx *fasthttp.RequestCtx) {
	var post models.Post
	if err := srv.ReadBody(ctx, &post); err != nil {
		srv.WriteError(ctx, srv.invalidFormatErr)
		return
	}

	slug := srv.readSlugOrID(ctx)
	id, err := strconv.ParseInt(slug, 10, 32)
	if err == nil {
		post.Thread = int32(id)
	} else {
		if err := srv.components.ThreadRepository.FindThreadIDBySlug(&post.Thread, slug); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	}

	if err := srv.components.PostValidator.Validate(&post); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	if err := srv.components.PostRepository.CreatePost(&post); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	srv.WriteJSON(ctx, http.StatusCreated, &post)
}
