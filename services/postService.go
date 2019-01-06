package services

import (
	"database/sql"
	"github.com/go-openapi/strfmt"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tp-project-db/consts"
	"tp-project-db/models"
	"tp-project-db/repositories"
)

func (srv *Server) createPost(ctx *fasthttp.RequestCtx) {
	var threadID int32
	threadSlug := srv.readSlugOrID(ctx)

	if id, err := strconv.ParseInt(threadSlug, 10, 32); err == nil {
		threadID = int32(id)
		if err := srv.components.ThreadRepository.CheckThreadExists(threadID); err != nil {
			srv.WriteError(ctx, err)
			return
		}
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

	attrs := strings.Split(string(ctx.QueryArgs().Peek("related")), ",")
	for _, attr := range attrs {
		switch string(attr) {
		case "forum":
			var forum models.Forum
			postMap["forum"] = &forum
		case "thread":
			var thread models.Thread
			postMap["thread"] = &thread
		case "user":
			var user models.User
			postMap["author"] = &user
		}
	}

	postPtr := (*models.PostFull)(&postMap)
	if err := srv.components.PostRepository.FindFullPost(postPtr); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, postPtr)
}

func (srv *Server) findPostsByThread(ctx *fasthttp.RequestCtx) {
	slugOrID := srv.readSlugOrID(ctx)

	sortType := string(ctx.QueryArgs().Peek("sort"))
	if sortType == consts.EmptyString {
		sortType = "flat"
	}

	since, parErr := ctx.QueryArgs().GetUint("since")
	if parErr != nil {
		since = 0
	}

	searchArgs := repositories.PostsByThreadSearchArgs{
		ThreadSlug: slugOrID,
		Since:      since,
		Limit:      srv.readLimit(ctx),
		Desc:       srv.readDescFlag(ctx),
		SortType:   sortType,
	}

	if id, err := strconv.ParseInt(slugOrID, 10, 32); err == nil {
		searchArgs.ThreadID = sql.NullInt64{
			Valid: true, Int64: id,
		}
	}

	log.Println(searchArgs)

	posts, err := srv.components.PostRepository.FindPostsByThread(&searchArgs)
	if err != nil {
		srv.WriteError(ctx, err)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, posts)
}

func (srv *Server) updatePost(ctx *fasthttp.RequestCtx) {
	id, err := srv.readID(ctx)
	if err != nil {
		srv.WriteError(ctx, err)
		return
	}
	post := models.Post{
		ID: id,
	}

	var postUpdate models.PostUpdate
	if err := srv.ReadBody(ctx, &postUpdate); err != nil {
		srv.WriteError(ctx, err)
		return
	}
	if postUpdate.Message == consts.EmptyString {
		if err := srv.components.PostRepository.FindPost(&post); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	} else {
		post.Message = postUpdate.Message
		if err := srv.components.PostRepository.UpdatePost(&post); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	}

	srv.WriteJSON(ctx, http.StatusOK, &post)
}
