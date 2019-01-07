package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"tp-project-db/models"
)

func (srv *Server) getStatus(ctx *fasthttp.RequestCtx) {
	var status models.Status
	_ = srv.components.StatusRepository.GetStatus(&status)
	srv.WriteJSON(ctx, http.StatusOK, &status)
}
