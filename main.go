package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
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
	handleErr(conn.Init())

	userRepository := repositories.NewUserRepository(conn)
	handleErr(userRepository.Init())

	forumRepository := repositories.NewForumRepository(conn)
	handleErr(forumRepository.Init())

	threadRepository := repositories.NewThreadRepository(conn)
	handleErr(threadRepository.Init())

	srv := services.NewServer(
		services.ServerConfig{
			Host: os.Getenv("SERVER_HOST"),
			Port: os.Getenv("SERVER_PORT"),
		},
		services.ServerComponents{
			UserValidator:       models.NewUserValidator(),
			UserUpdateValidator: models.NewUserUpdateValidator(),
			UserRepository:      userRepository,

			ForumValidator:  models.NewForumValidator(),
			ForumRepository: forumRepository,

			ThreadValidator:       models.NewThreadValidator(),
			ThreadUpdateValidator: models.NewThreadUpdateValidator(),
			ThreadRepository:      threadRepository,
		},
	)

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)

	go func() {
		handleErr(srv.Run())
		ch <- os.Kill
	}()

	log.Println("server started...")
	<-ch

	handleErr(srv.Shutdown())
}

func handleErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v\n%s", err, string(debug.Stack())))
	}
}
