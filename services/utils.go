package services

import (
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
	"tp-project-db/errs"
)

func (srv *Server) ReadBody(ctx *fasthttp.RequestCtx, v easyjson.Unmarshaler) *errs.Error {
	if err := easyjson.Unmarshal(ctx.PostBody(), v); err != nil {
		return srv.invalidFormatErr
	}
	return nil
}

func (srv *Server) WriteJSON(ctx *fasthttp.RequestCtx, status int, v easyjson.Marshaler) {

}

func (srv *Server) WriteError(ctx *fasthttp.RequestCtx, err *errs.Error) {

}

func (srv *Server) readNickname(ctx *fasthttp.RequestCtx) string {
	return ctx.UserValue("nickname").(string)
}
