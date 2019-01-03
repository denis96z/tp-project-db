package services

import (
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
	"tp-project-db/errs"
)

const (
	JsonType = "application/json"
)

func (srv *Server) ReadBody(ctx *fasthttp.RequestCtx, v easyjson.Unmarshaler) *errs.Error {
	if err := easyjson.Unmarshal(ctx.PostBody(), v); err != nil {
		return srv.invalidFormatErr
	}
	return nil
}

func (srv *Server) WriteJSON(ctx *fasthttp.RequestCtx, status int, v easyjson.Marshaler) {
	b, _ := easyjson.Marshal(v)
	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody(b)
}

func (srv *Server) WriteError(ctx *fasthttp.RequestCtx, err *errs.Error) {
	b, _ := easyjson.Marshal(err)
	ctx.SetStatusCode(err.HttpStatus)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody(b)
}

func (srv *Server) readNickname(ctx *fasthttp.RequestCtx) string {
	return ctx.UserValue("nickname").(string)
}
