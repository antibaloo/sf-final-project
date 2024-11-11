package main

import (
	"fmt"
	"time"

	"github.com/antibaloo/sf-final-project/internal/api/censor"
	"github.com/antibaloo/sf-final-project/internal/config"
	"github.com/antibaloo/sf-final-project/internal/storage/postgres"
)

func main() {
	// Читаем конфигурацию
	config, err := config.NewConfig()
	if err != nil {
		fmt.Printf("%v: ошибка при загрузке конфигурации сервиса проверки комментариев: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
	// Подключаемся к БД
	db, err := postgres.NewStore(config.ConString())
	if err != nil {
		fmt.Printf("%v: ошибка при соединении с БД: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
	// Создаем сервис проверки комментариев
	censorServer, err := censor.CreateService(config.CensorAddress(), db)
	if err != nil {
		fmt.Printf("%v: ошибка при создании сервиса проверки комментариев: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
	// Запускаем сорвис проверки комментариев
	if err := censorServer.Start(); err != nil {
		fmt.Printf("%v: ошибка при запуске сервиса проверки комментариев: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
}
