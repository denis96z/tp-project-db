package services

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"tp-project-db/errs"
	"tp-project-db/models"
	"tp-project-db/repositories"
)

const (
	InvalidFormatErrMessage = "invalid request parameter"
)

type ServerConfig struct {
	Host string
	Port string
}

type ServerComponents struct {
	UserValidator       *models.UserValidator
	UserUpdateValidator *models.UserUpdateValidator
	UserRepository      *repositories.UserRepository

	ForumValidator  *models.ForumValidator
	ForumRepository *repositories.ForumRepository

	ThreadValidator       *models.ThreadValidator
	ThreadUpdateValidator *models.ThreadUpdateValidator
	ThreadRepository      *repositories.ThreadRepository

	PostValidator  *models.PostValidator
	PostRepository *repositories.PostRepository
}

type Server struct {
	handler fasthttp.RequestHandler

	config     ServerConfig
	components ServerComponents

	invalidFormatErr *errs.Error
}

func NewServer(config ServerConfig, components ServerComponents) *Server {
	srv := &Server{
		config:           config,
		components:       components,
		invalidFormatErr: errs.NewInvalidFormatError(InvalidFormatErrMessage),
	}

	r := router.New()

	r.POST("/api/forum/:slug/create", srv.createThread)
	r.GET("/api/forum/:slug/details", srv.findForumBySlug)
	r.GET("/api/forum/:slug/threads", srv.findThreadsByForum)
	r.POST("/api/thread/:slug_or_id/create", srv.createPost)
	r.GET("/api/thread/:slug_or_id/details", srv.findThreadBySlugOrID)
	r.POST("/api/user/:nickname/create", srv.createUser)
	r.GET("/api/user/:nickname/profile", srv.findUserByNickname)
	r.POST("/api/user/:nickname/profile", srv.updateUserByNickname)
	r.POST("/api/service/clear", srv.deleteAllUsers)

	srv.handler = func(r *router.Router) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			if string(ctx.Path()) == "/api/forum/create" {
				srv.createForum(ctx)
				return
			}
			r.Handler(ctx)
		}
	}(r)
	return srv
}

func (srv *Server) Run() error {
	addr := ":" + srv.config.Port
	return fasthttp.ListenAndServe(addr, srv.handler)
}

func (srv *Server) Shutdown() error {
	return nil //TODO
}
