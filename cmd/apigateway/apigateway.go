package main

import (
	"fmt"
	"time"

	"github.com/antibaloo/sf-final-project/internal/api/gateway"
	"github.com/antibaloo/sf-final-project/internal/config"
)

func main() {
	// Читаем конфигурацию
	config, err := config.NewConfig()
	if err != nil {
		fmt.Printf("%v: ошибка при загрузке конфигурации сервиса шлюза: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
	// Создаем сервис шлюза
	apiGateway, err := gateway.CreateService(
		config.APIGatewayAddress(),
		config.NewsAddress(),
		config.CommentsAddress(),
		config.CensorAddress(),
	)
	if err != nil {
		fmt.Printf("%v: ошибка при создании сервиса шлюза: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		return
	}
	// Запускаем сервис шлюза
	if err := apiGateway.Start(); err != nil {
		fmt.Printf("%v: ошибка во время запуска сервиса шлюза: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err)
		return
	}
}
