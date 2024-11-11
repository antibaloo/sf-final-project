package comments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/antibaloo/sf-final-project/internal/api/middleware"
	"github.com/antibaloo/sf-final-project/internal/storage"
)

// Структура сервиса комментариев
type commentsService struct {
	address    string               // адрес на котором будет запущен сервис
	modCh      chan storage.Comment // канала связи с горутиной модератора комментариев
	db         storage.Store        // база данных сервиса
	httpServer *http.Server         // веб-сервер сервиса
}

// Конструктор структуры сервиса комментариев
func CreateService(address string, db storage.Store) (*commentsService, error) {
	if address == "" {
		return nil, fmt.Errorf("адрес сервиса отсутствует")
	}
	if db == nil {
		return nil, fmt.Errorf("указательна структуру БД отсутствует")
	}
	return &commentsService{
		address: address,
		db:      db,
		modCh:   make(chan storage.Comment),
	}, nil
}

// Метод запуска сервиса комментариев
func (comments *commentsService) Start() error {
	fmt.Printf("%v: запускаем сервис комментариев по адресу: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), comments.address)
	router := http.NewServeMux()
	router.HandleFunc("GET /news/{id}/comments", comments.getCommentsByNewsIdHandler)
	router.HandleFunc("POST /comment", comments.addCommentHandler)
	comments.httpServer = &http.Server{
		Addr:    comments.address,
		Handler: middleware.GenIdAndLogging(router),
	}
	// create channel to listen for signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := comments.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("%v: ошибка при запуске сервиса комментариев: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		}
	}()

	s := <-stopChan
	fmt.Printf("%v: получен сигнал: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), s.String())
	if err := comments.httpServer.Shutdown(context.Background()); err != nil {
		return err
	}
	return nil
}

// Обработчик получения списка комментариев к новости
func (comments *commentsService) getCommentsByNewsIdHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	commentsByNewsId, err := comments.db.CommentsByNewsId(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bytes, err := json.Marshal(commentsByNewsId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// Обработчик добавления нового комментария
func (comments *commentsService) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	var comment storage.Comment
	// Декодируем тело запроса и проверяемна ошибки
	err := json.NewDecoder(r.Body).Decode(&comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Добавляем комментарий в БД и проверяем на ошибки
	err = comments.db.AddComment(comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
