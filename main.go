package main

import (
	"os"
	"tp-project-db/models"
	"tp-project-db/repositories"
	"tp-project-db/services"
)

func main() {
	_ = os.Setenv("PGHOST", "127.0.0.1")
	_ = os.Setenv("PGPORT", "5432")
	_ = os.Setenv("PGUSER", "user")
	_ = os.Setenv("PGPASSWORD", "password")
	_ = os.Setenv("PGDATABASE", "forum")

	conn := repositories.NewConnection()
	handleErr(conn.Open())
	defer func() {
		handleErr(conn.Close())
	}()

	userRepository := repositories.NewUserRepository(conn)
	handleErr(userRepository.Init())

	srv := services.NewServer(
		services.ServerConfig{
			Port: "5000",
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
		panic(err)
	}
}
