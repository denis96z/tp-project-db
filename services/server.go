package services

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type ServerConfig struct {
	Host string
	Port string
}

type ServerComponents struct {
}

type Server struct {
	router     *router.Router
	config     ServerConfig
	components ServerComponents
}

func NewServer(config ServerConfig, components ServerComponents) *Server {
	srv := &Server{
		config:     config,
		components: components,
	}

	r := router.New()

	r.POST("/api/user/:nickname/create", srv.createUser)

	srv.router = r
	return srv
}

func (srv *Server) Run() error {
	addr := ":" + srv.config.Port
	return fasthttp.ListenAndServe(addr, srv.router.Handler)
}
