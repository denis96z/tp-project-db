package services

import (
	"github.com/mailru/easyjson"
	"tp-project-db/errs"
)

func (srv *Server) WriteOk(status int, v easyjson.Marshaler) {

}

func (srv *Server) WriteError(err *errs.Error) {

}
