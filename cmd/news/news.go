package main

import (
	"fmt"
	"time"

	"github.com/antibaloo/sf-final-project/internal/api/news"
	"github.com/antibaloo/sf-final-project/internal/config"
	"github.com/antibaloo/sf-final-project/internal/rss"
	"github.com/antibaloo/sf-final-project/internal/storage/postgres"
)

func main() {
	config, err := config.NewConfig()
	if err != nil {
		fmt.Printf("%v: ошибка при загрузке кофигурации сервиса новостей: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}

	db, err := postgres.NewStore(config.ConString())
	if err != nil {
		fmt.Printf("%v: ошбика при боключении сервиса новостей к БД: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
	defer db.Pool.Close()

	// Создаем ридер новостей
	rssReader, err := rss.CreateService(config.RssConfig(), db)
	if err != nil {
		fmt.Printf("%v: ошибка при создании rss ридера новостей: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}

	// Запускаем ридер новостей
	rssReader.Start()

	// Создаем сервис новостей
	newsServer, err := news.CreateService(config.NewsAddress(), config.NewsPerPage(), db)
	if err != nil {
		fmt.Printf("%v: ошибка при создании сервиса новостей: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
	// Запускаем сервис новостей
	if err := newsServer.Start(); err != nil {
		fmt.Printf("%v: ошибка при запуске сервиса новостей: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err)
		return
	}
}
