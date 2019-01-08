package services

import (
	"database/sql"
	"github.com/go-openapi/strfmt"
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"strings"
	"time"
	"tp-project-db/consts"
	"tp-project-db/models"
	"tp-project-db/repositories"
)

func (srv *Server) createPosts(ctx *fasthttp.RequestCtx) {
	args := repositories.CreatePostArgs{
		ThreadID:  -1,
		Timestamp: strfmt.DateTime(time.Now()),
	}
	args.ThreadSlug = ctx.UserValue("slug_or_id").(string)

	if id, err := strconv.ParseInt(args.ThreadSlug, 10, 32); err == nil {
		args.ThreadID = int32(id)
		if srv.components.ThreadRepository.FindThreadForumByID(&args) != nil {
			srv.WriteError(ctx, http.StatusNotFound)
			return
		}
	} else {
		if srv.components.ThreadRepository.FindThreadIDAndForumBySlug(&args) != nil {
			srv.WriteError(ctx, http.StatusNotFound)
			return
		}
	}

	var posts models.Posts
	srv.ReadBody(ctx, &posts)

	n := len(([]models.Post)(posts))
	if n == 0 {
		srv.WriteJSON(ctx, http.StatusCreated, &posts)
		return
	}

	if err := srv.components.PostRepository.CreatePosts(&posts, &args); err != nil {
		srv.WriteError(ctx, err.HttpStatus)
		return
	}

	srv.rwMtx.Lock()
	srv.status.NumPosts += int64(n)
	srv.rwMtx.Unlock()

	srv.WriteJSON(ctx, http.StatusCreated, &posts)
}

func (srv *Server) findPost(ctx *fasthttp.RequestCtx) {
	id, _ := strconv.ParseInt(ctx.UserValue("id").(string), 10, 64)
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
		srv.WriteError(ctx, err.HttpStatus)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, postPtr)
}

func (srv *Server) findPostsByThread(ctx *fasthttp.RequestCtx) {
	slugOrID := ctx.UserValue("slug_or_id").(string)

	sortType := string(ctx.QueryArgs().Peek("sort"))
	if sortType == consts.EmptyString {
		sortType = "flat"
	}

	since, parErr := ctx.QueryArgs().GetUint("since")
	if parErr != nil {
		since = 0
	}

	limit, parErr := ctx.QueryArgs().GetUint("limit")
	if parErr != nil {
		since = 0
	}

	searchArgs := repositories.PostsByThreadSearchArgs{
		ThreadSlug: slugOrID,
		Since:      since,
		Limit:      limit,
		Desc:       ctx.QueryArgs().GetBool("desc"),
		SortType:   sortType,
	}

	if id, err := strconv.ParseInt(slugOrID, 10, 32); err == nil {
		searchArgs.ThreadID = sql.NullInt64{
			Valid: true, Int64: id,
		}
	}

	posts, err := srv.components.PostRepository.FindPostsByThread(&searchArgs)
	if err != nil {
		srv.WriteError(ctx, err.HttpStatus)
		return
	}

	srv.WriteJSON(ctx, http.StatusOK, posts)
}
