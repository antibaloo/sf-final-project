package censor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/antibaloo/sf-final-project/internal/api/middleware"
	"github.com/antibaloo/sf-final-project/internal/storage"
)

// Структура сервиса проверки комментариев
type censor struct {
	dictionary []string      // Словать запрещенных слов
	address    string        // адрес на котором будет запущен сервис
	db         storage.Store // база данных сервиса
	httpServer *http.Server  // веб-сервер сервиса
}

// Конструктор сервиса проверки комментариев
func CreateService(address string, db storage.Store) (*censor, error) {
	if address == "" {
		return nil, fmt.Errorf("не указан адрес запуска сервиса")
	}
	if db == nil {
		return nil, fmt.Errorf("не указана база данных")
	}
	return &censor{
		dictionary: []string{},
		address:    address,
		db:         db,
	}, nil
}

// Метод запуска сервиса проверки комментариев
func (censor *censor) Start() error {
	fmt.Printf("%v: запускаем сервис комментариев по адресу: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), censor.address)

	// Загружаем словарь из БД
	dictionary, err := censor.db.Dictionary()
	if err != nil {
		return err
	}
	fmt.Printf("%v: загружен словать запрещенных слов: %v\n", time.Now().Format("02.01.2006 15:04:05 MST"), dictionary)
	censor.dictionary = dictionary
	router := http.NewServeMux()
	router.HandleFunc("POST /check", censor.checkHandler)
	censor.httpServer = &http.Server{
		Addr:    censor.address,
		Handler: middleware.GenIdAndLogging(router),
	}

	// создаем канал для сигналов
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		if err := censor.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("%v: ошибка при запуске сервиса: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		}
	}()

	// Читаем из канала
	s := <-stopChan
	fmt.Printf("%v: получен сигнал: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), s.String())
	if err := censor.httpServer.Shutdown(context.Background()); err != nil {
		return err
	}
	return nil
}

// Обработчик проверки комментария
func (censor *censor) checkHandler(w http.ResponseWriter, r *http.Request) {
	var comment storage.Comment
	// Декодируем пэйлоад запроса и проверяемна ошибки
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, word := range censor.dictionary {
		if strings.Contains(strings.ToLower(comment.Content), word) {
			http.Error(w, "комментарий содержит запрещенное слово", http.StatusBadRequest)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}
