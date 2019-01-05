package services

import (
	"github.com/go-openapi/strfmt"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"time"
	"tp-project-db/models"
)

func (srv *Server) createPost(ctx *fasthttp.RequestCtx) {
	var threadID int32
	threadSlug := srv.readSlugOrID(ctx)

	if id, err := strconv.ParseInt(threadSlug, 10, 32); err == nil {
		threadID = int32(id)
	} else {
		if err := srv.components.ThreadRepository.FindThreadIDBySlug(&threadID, threadSlug); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	}

	var posts models.Posts
	if err := srv.ReadBody(ctx, &posts); err != nil {
		srv.WriteError(ctx, srv.invalidFormatErr)
		return
	}

	currentTimestamp := models.NullTimestamp{
		Valid:     true,
		Timestamp: strfmt.DateTime(time.Now()),
	}

	n := len(posts)
	if n == 0 {
		srv.WriteJSON(ctx, http.StatusNotFound, srv.invalidFormatErr)
		return
	}

	for i := 0; i < n; i++ {
		posts[i].Thread = threadID
		posts[i].CreatedTimestamp = currentTimestamp

		if err := srv.components.PostValidator.Validate(&posts[i]); err != nil {
			srv.WriteError(ctx, err)
			return
		}

		if err := srv.components.PostRepository.CreatePost(&posts[i]); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	}

	srv.WriteJSON(ctx, http.StatusCreated, &posts)
}
