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

	if err := srv.components.ThreadRepository.CheckThreadExists(threadID); err != nil {
		srv.WriteError(ctx, err)
		return
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

	for i := 0; i < len(posts); i++ {
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

func (srv *Server) findPost(ctx *fasthttp.RequestCtx) {
	id, err := srv.readID(ctx)
	if err != nil {
		srv.WriteError(ctx, err)
		return
	}

	postMap := make(map[string]interface{}, 0)

	post := models.Post{
		ID: id,
	}
	postMap["post"] = &post

	attrs := ctx.QueryArgs().PeekMulti("related")
	for _, attr := range attrs {
		switch string(attr) {
		case "forum":
			var forum models.Forum
			postMap["forum"] = &forum
		case "thread":
			var thread models.Thread
			postMap["thread"] = &thread
		case "author":
			var user models.User
			postMap["user"] = &user
		}
	}

	postPtr := (*models.PostFull)(&postMap)
	if err := srv.components.PostRepository.FindPostByID(postPtr); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, postPtr)
}
