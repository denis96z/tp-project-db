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
}

type Server struct {
	router *router.Router

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

	r.POST("/api/user/:nickname/create", srv.createUser)
	r.GET("/api/user/:nickname/profile", srv.findUserByNickname)
	r.POST("/api/user/:nickname/profile", srv.updateUserByNickname)

	srv.router = r
	return srv
}

func (srv *Server) Run() error {
	addr := ":" + srv.config.Port
	return fasthttp.ListenAndServe(addr, srv.router.Handler)
}

func (srv *Server) Shutdown() error {
	return nil //TODO
}
