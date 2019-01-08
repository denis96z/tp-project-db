package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
)

func (srv *Server) getStatus(ctx *fasthttp.RequestCtx) {
	srv.rwMtx.RLock()
	status := srv.status
	srv.rwMtx.RUnlock()
	srv.WriteJSON(ctx, http.StatusOK, &status)
}

func (srv *Server) clearDatabase(ctx *fasthttp.RequestCtx) {
	srv.components.StatusRepository.ClearDatabase()

	srv.rwMtx.RLock()
	srv.status.NumUsers = 0
	srv.status.NumForums = 0
	srv.status.NumThreads = 0
	srv.status.NumPosts = 0
	srv.rwMtx.RUnlock()
}
