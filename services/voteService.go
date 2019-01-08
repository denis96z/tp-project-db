package services

import (
	"database/sql"
	"github.com/valyala/fasthttp"
	"strconv"
	"tp-project-db/models"
)

func (srv *Server) addVote(ctx *fasthttp.RequestCtx) {
	var vote models.Vote
	srv.ReadBody(ctx, &vote)

	vote.ThreadSlug = ctx.UserValue("slug_or_id").(string)

	id, err := strconv.ParseInt(vote.ThreadSlug, 10, 32)
	if err == nil {
		vote.ThreadID = int32(id)
	}

	var thread sql.NullString
	status := srv.components.VoteRepository.AddVote(&vote, &thread)

	if thread.Valid {
		ctx.SetStatusCode(status)
		ctx.Response.Header.SetContentType(JsonType)
		ctx.Response.SetBody([]byte(thread.String))
	} else {
		srv.WriteError(ctx, status)
	}
}
