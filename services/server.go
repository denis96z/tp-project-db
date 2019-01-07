package services

import (
	"github.com/fasthttp/router"
	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
	"net/http"
	"sync"
	"tp-project-db/errs"
	"tp-project-db/models"
	"tp-project-db/repositories"
)

type ServerConfig struct {
	Host string
	Port string
}

type ServerComponents struct {
	UserRepository   *repositories.UserRepository
	ForumRepository  *repositories.ForumRepository
	ThreadRepository *repositories.ThreadRepository
	PostRepository   *repositories.PostRepository
	StatusRepository *repositories.StatusRepository
}

type Server struct {
	handler fasthttp.RequestHandler

	config     ServerConfig
	components ServerComponents

	rwMtx  *sync.RWMutex
	status models.Status

	commonErr []byte
}

func NewServer(config ServerConfig, components ServerComponents) *Server {
	srv := &Server{
		config:     config,
		components: components,

		status: models.Status{},
		rwMtx:  &sync.RWMutex{},

		commonErr: func() []byte {
			err := errs.NewError(http.StatusInternalServerError, "error")
			b, _ := easyjson.Marshal(err)
			return b
		}(),
	}

	r := router.New()
	components.StatusRepository.GetStatus(&srv.status)

	/*r.GET("/debug/pprof/profile", fasthttpadaptor.NewFastHTTPHandlerFunc(pprof.Profile))*/
	r.POST("/api/forum/:slug/create", srv.createThread)
	r.GET("/api/forum/:slug/details", srv.findForum) /*
		r.GET("/api/forum/:slug/threads", srv.findThreadsByForum)
		r.GET("/api/forum/:slug/users", srv.findUsersByForum)
		r.GET("/api/post/:id/details", srv.findPost)
		r.POST("/api/post/:id/details", srv.updatePost)*/
	r.POST("/api/thread/:slug_or_id/create", srv.createPosts) /*
		r.POST("/api/thread/:slug_or_id/vote", srv.addVote)*/
	r.GET("/api/thread/:slug_or_id/details", srv.findThread)
	r.GET("/api/thread/:slug_or_id/posts", srv.findPostsByThread) /*
		r.POST("/api/thread/:slug_or_id/details", srv.updateThread)*/
	r.POST("/api/user/:nickname/create", srv.createUser)
	r.GET("/api/user/:nickname/profile", srv.findUser)
	r.POST("/api/user/:nickname/profile", srv.updateUser)
	/*r.POST("/api/service/clear", srv.deleteAllUsers)*/
	r.GET("/api/service/status", srv.getStatus)

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
