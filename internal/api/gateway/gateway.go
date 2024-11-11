package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/antibaloo/sf-final-project/internal/api/middleware"
	"github.com/antibaloo/sf-final-project/internal/storage"
)

// Структура apiGateway
type apiGateway struct {
	address        string
	newsAddress    string
	commentsAddres string
	censorAddress  string
	httpServer     *http.Server
}

// Конструктор сервиса apigateway
func CreateService(address, newsAddress, commentsAddress, censorAddress string) (*apiGateway, error) {
	if address == "" {
		return nil, fmt.Errorf("адрес сервиса отсутствует")
	}
	if newsAddress == "" {
		return nil, fmt.Errorf("адрес сервиса новостей отсутствует")
	}
	if commentsAddress == "" {
		return nil, fmt.Errorf("адрес сервиса комментариев отсутствует")
	}
	if censorAddress == "" {
		return nil, fmt.Errorf("адрес сервиса проверки комментариев отсутствует")
	}
	return &apiGateway{
		address:        address,
		newsAddress:    newsAddress,
		commentsAddres: commentsAddress,
		censorAddress:  censorAddress}, nil
}

// Метод запуска сервиса шлюза
func (api *apiGateway) Start() error {
	fmt.Printf("%v: запускаем apiGateway по адресу: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), api.address)
	router := http.NewServeMux()
	router.HandleFunc("GET /news", api.newsHandler)
	router.HandleFunc("GET /news/{id}", api.detailedNewsHandler)
	router.HandleFunc("POST /comment", api.addCommentHandler)
	api.httpServer = &http.Server{
		Addr:    api.address,
		Handler: middleware.GenIdAndLogging(router),
	}
	// create channel to listen for signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := api.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("%v: ошибка при запуске apiGateway: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		}
	}()

	s := <-stopChan
	fmt.Printf("%v: получен сигнал: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), s.String())
	if err := api.httpServer.Shutdown(context.Background()); err != nil {
		return err
	}
	return nil
}

// Обработчик получения списка новостей
func (api *apiGateway) newsHandler(w http.ResponseWriter, r *http.Request) {
	// Перенаправляем запрос по адресу сервиса новостей
	resp, err := http.Get("http://" + api.newsAddress + r.URL.RequestURI())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Возвращаем клиенту код и тело ответа от сервиса новостей
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// Обработчик получения детальной новости с комментариями
func (api *apiGateway) detailedNewsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		respNews, respComments *http.Response
		errNews, errComments   error
		wg                     sync.WaitGroup
		news                   storage.NewsFullDetailed
		comments               []storage.Comment
	)
	wg.Add(2)
	// Запускаем ассинхронно запросы к сервисам новостей и комментариев
	go func() {
		defer wg.Done()
		respNews, errNews = http.Get("http://" + api.newsAddress + r.URL.Path + "/detailed?" + r.URL.RawQuery)
	}()
	go func() {
		defer wg.Done()
		respComments, errComments = http.Get("http://" + api.commentsAddres + r.URL.Path + "/comments?" + r.URL.RawQuery)
	}()

	// Ждем пока отработают оба запроса
	wg.Wait()

	// Проверяем ошибку в запросе к новостям
	if errNews != nil {
		http.Error(w, errNews.Error(), http.StatusInternalServerError)
		return
	}

	// Если сервис новостей вернул ошибку
	if respNews.StatusCode != http.StatusOK {
		defer respNews.Body.Close()
		// Читаем тело ответа
		newsBody, err := io.ReadAll(respNews.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Возвращаем клиенту код и тело ответа от сервиса новостей
		w.WriteHeader(respNews.StatusCode)
		w.Write(newsBody)
		return
	}

	// Проверяем ошибку в запросе к комментариям
	if errComments != nil {
		http.Error(w, errNews.Error(), http.StatusInternalServerError)
		return
	}

	// Если сервис комментариев вернул ошибку
	if respComments.StatusCode != http.StatusOK {
		defer respNews.Body.Close()
		// Читаем тело ответа
		commentsBody, err := io.ReadAll(respComments.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Возвращаем клиенту код и тело ответа от сервиса новостей
		w.WriteHeader(respComments.StatusCode)
		w.Write(commentsBody)
		return
	}

	// Раскодируем тело ответа новостей в структуру
	err := json.NewDecoder(respNews.Body).Decode(&news)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Раскодируем тело комментариев в массив структур
	err = json.NewDecoder(respComments.Body).Decode(&comments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Объединяем новости с комментариями
	news.Comments = comments
	// Кодируем в json
	bytes, err := json.Marshal(news)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Отдаем клиенту
	w.Write(bytes)
}

// Обработчик добавления нового комментария
func (api *apiGateway) addCommentHandler(w http.ResponseWriter, r *http.Request) {
	// Сохраняем тело запроса, чтобы отправить его нескольким получателям
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Отправляем полуяенный комментрий на проверку к сервису проверки
	resp, err := http.Post("http://"+api.censorAddress+"/check?"+r.URL.RawQuery, "application/json", bytes.NewReader(body))
	// Проверяем на ошибку запрос к сервису проверки комментариев
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Проверяем ответ сервиса проверки комментариев
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		// Читаем тело ответа
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Возвращаем клиенту код и тело ответа от сервиса новостей
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	// Если проверка пройдена, отправляем комментарий на публикацию
	resp, err = http.Post("http://"+api.commentsAddres+r.URL.Path+"?"+r.URL.RawQuery, "application/json", bytes.NewReader(body))
	// Проверяем на ошибку запрос к сервису комментариев
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()
	// Читаем тело ответа
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Возвращаем клиенту код и тело ответа от сервиса комментариев
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}
