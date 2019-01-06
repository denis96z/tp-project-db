package services

import (
	"github.com/valyala/fasthttp"
	"net/http"
	"tp-project-db/models"
)

func (srv *Server) getStatus(ctx *fasthttp.RequestCtx) {
	var status models.Status
	if err := srv.components.StatusRepository.GetStatus(&status); err != nil {
		srv.WriteError(ctx, err)
		return
	}
	srv.WriteJSON(ctx, http.StatusOK, &status)
}
