package services

import (
	"github.com/go-openapi/strfmt"
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
	"strconv"
	"tp-project-db/errs"
)

const (
	JsonType = "application/json"
)

func (srv *Server) ReadBody(ctx *fasthttp.RequestCtx, v easyjson.Unmarshaler) *errs.Error {
	if easyjson.Unmarshal(ctx.PostBody(), v) != nil {
		return srv.invalidFormatErr
	}
	return nil
}

func (srv *Server) WriteCommonError(ctx *fasthttp.RequestCtx, status int) {
	ctx.SetStatusCode(status)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody(srv.commonErr)
}

func (srv *Server) ReadBodyAllowEmpty(ctx *fasthttp.RequestCtx, v easyjson.Unmarshaler) *errs.Error {
	b := ctx.PostBody()
	if len(b) == 2 {
		return srv.invalidFormatErr
	}
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
	/*b, _ := easyjson.Marshal(err)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody(b)*/
	ctx.SetStatusCode(err.HttpStatus)
	ctx.Response.Header.SetContentType(JsonType)
	ctx.Response.SetBody(srv.commonErr)
}

func (srv *Server) readNickname(ctx *fasthttp.RequestCtx) string {
	return srv.readPathParam(ctx, "nickname")
}

func (srv *Server) readID(ctx *fasthttp.RequestCtx) (int64, *errs.Error) {
	idStr := ctx.UserValue("id").(string)
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		return id, nil
	}
	return 0, srv.invalidFormatErr
}

func (srv *Server) readSlug(ctx *fasthttp.RequestCtx) string {
	return srv.readPathParam(ctx, "slug")
}

func (srv *Server) readSlugOrID(ctx *fasthttp.RequestCtx) string {
	return srv.readPathParam(ctx, "slug_or_id")
}

func (srv *Server) readDescFlag(ctx *fasthttp.RequestCtx) bool {
	return ctx.QueryArgs().GetBool("desc")
}

func (srv *Server) readSinceNickname(ctx *fasthttp.RequestCtx) string {
	return string(ctx.QueryArgs().Peek("since"))
}

func (srv *Server) readSinceTimestamp(ctx *fasthttp.RequestCtx, since *strfmt.DateTime) error {
	value := ctx.QueryArgs().Peek("since")
	if len(value) == 0 {
		return srv.invalidFormatErr
	}
	return since.UnmarshalText(ctx.QueryArgs().Peek("since"))
}

func (srv *Server) readLimit(ctx *fasthttp.RequestCtx) int {
	return ctx.QueryArgs().GetUintOrZero("limit")
}

func (srv *Server) readPathParam(ctx *fasthttp.RequestCtx, name string) string {
	return ctx.UserValue(name).(string)
}
