package main

import (
	"fmt"
	"time"

	"github.com/antibaloo/sf-final-project/internal/api/comments"
	"github.com/antibaloo/sf-final-project/internal/config"
	"github.com/antibaloo/sf-final-project/internal/storage/postgres"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		fmt.Printf("%v: ошибка при запуске сервиса комментариев: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
	db, err := postgres.NewStore(config.ConString())
	if err != nil {
		fmt.Printf("%v: ошибка при соединении сервиса комменатриев с БД: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}

	commentsServer, err := comments.CreateService(config.CommentsAddress(), db)
	if err != nil {
		fmt.Printf("%v: ошибка при создании сервиса комменатриев: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}

	if err := commentsServer.Start(); err != nil {
		fmt.Printf("%v: ошибка при запуске сервиса комментариев: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
}
