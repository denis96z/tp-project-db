package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"strconv"
	"tp-project-db/models"
)

func (srv *Server) addVote(ctx *fasthttp.RequestCtx) {
	var vote models.Vote
	if err := srv.ReadBody(ctx, &vote); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	if err := srv.components.VoteValidator.Validate(&vote); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	threadSlug := srv.readSlugOrID(ctx)
	id, err := strconv.ParseInt(threadSlug, 10, 32)
	if err == nil {
		vote.Thread = int32(id)
	} else {
		if err := srv.components.ThreadRepository.FindThreadIDBySlug(&vote.Thread, threadSlug); err != nil {
			srv.WriteError(ctx, err)
			return
		}
	}

	if err := srv.components.VoteRepository.AddVote(&vote); err != nil {
		srv.WriteError(ctx, err)
		return
	}

	thread := models.Thread{
		ID: vote.Thread,
	}
	if err := srv.components.ThreadRepository.FindThreadByID(&thread); err != nil {
		panic(err)
	}

	srv.WriteJSON(ctx, http.StatusOK, &thread)
}
