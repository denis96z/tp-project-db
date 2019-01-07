package services

import (
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

const (
	JsonType = "application/json"
)

func (srv *Server) ReadBody(ctx *fasthttp.RequestCtx, v easyjson.Unmarshaler) {
	_ = easyjson.Unmarshal(ctx.PostBody(), v)
}

func (srv *Server) WriteJSON(ctx *fasthttp.RequestCtx, status int, v easyjson.Marshaler) {
	b, _ := easyjson.Marshal(v)
	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody(b)
}

func (srv *Server) WriteError(ctx *fasthttp.RequestCtx, status int) {
	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody(srv.commonErr)
}
