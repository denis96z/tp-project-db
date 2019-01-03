package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"tp-project-db/config"
	"tp-project-db/models"
	"tp-project-db/repositories"
	"tp-project-db/services"
)

func main() {
	config.Load()

	conn := repositories.NewConnection()
	handleErr(conn.Open())
	defer func() {
		handleErr(conn.Close())
	}()

	userRepository := repositories.NewUserRepository(conn)
	handleErr(userRepository.Init())

	srv := services.NewServer(
		services.ServerConfig{
			Host: os.Getenv("SERVER_HOST"),
			Port: os.Getenv("SERVER_PORT"),
		},
		services.ServerComponents{
			UserValidator:       models.NewUserValidator(),
			UserUpdateValidator: models.NewUserUpdateValidator(),
			UserRepository:      userRepository,
		},
	)
	handleErr(srv.Run())
	defer func() {
		handleErr(srv.Shutdown())
	}()
}

func handleErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v\n%s", err, string(debug.Stack())))
	}
}
