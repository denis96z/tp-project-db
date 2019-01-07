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
