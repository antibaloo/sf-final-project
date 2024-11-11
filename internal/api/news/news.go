package news

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

// Структура объекта паджинации
type pagination struct {
	NewsPerPage int `json:"news_per_page"` // Новостей на странице
	Page        int `json:"page"`          // Текущая страница
	Pages       int `json:"pages"`         // Всего страниц
}

// Структура для возвращения списка новостей с объектом паджинации
type newsResponse struct {
	News       []storage.NewsShortDetailed `json:"news"`
	Pagination pagination                  `json:"pagination"`
}

// Структура сервиса новостей
type newsService struct {
	address     string
	db          storage.Store
	httpServer  *http.Server
	newsPerPage int
}

// Конструктор структуры сервиса новостей
func CreateService(address string, n int, db storage.Store) (*newsService, error) {
	if address == "" {
		return nil, fmt.Errorf("адрес обязательный параметр")
	}
	if db == nil {
		return nil, fmt.Errorf("объек БД обязательный параметр")
	}

	return &newsService{
		address:     address,
		newsPerPage: n,
		db:          db,
	}, nil
}

func (news *newsService) Start() error {
	fmt.Printf("%v: запускаем сервис новостей по адресу: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), news.address)
	router := http.NewServeMux()
	router.HandleFunc("GET /news", news.newsHandler)
	router.HandleFunc("GET /news/{id}/detailed", news.detailedNewsHandler)
	news.httpServer = &http.Server{
		Addr:    news.address,
		Handler: middleware.GenIdAndLogging(router),
	}
	// create channel to listen for signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := news.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("%v: ошибка при запуске сервиса новостей: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), err.Error())
		}
	}()

	s := <-stopChan
	fmt.Printf("%v: получен сигнал: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), s.String())
	if err := news.httpServer.Shutdown(context.Background()); err != nil {
		return err
	}
	return nil
}

// Обработчик получения списка новостей
func (n *newsService) newsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		p            pagination     // Объект паджинации
		page         int        = 1 // Значение по-умолчанию
		offset       int
		err          error
		newsResponse newsResponse //Структура для возвращения списка новостей с объектом паджинации
	)
	// Читаем строку поиска
	search := r.URL.Query().Get("search")

	// Читаем номер страницы
	pageParam := r.URL.Query().Get("page")
	// Если параметр непустой, конвертируем его в число
	if pageParam != "" {
		// параметр page - это число, поэтому нужно сконвертировать
		// строку в число при помощи пакета strconv
		page, err = strconv.Atoi(pageParam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	// Инициализируем объект паджинации
	p.NewsPerPage = n.newsPerPage
	p.Page = page
	// Рассчитываем смещение, если страница не первая
	if page > 1 {
		offset = (page - 1) * n.newsPerPage
	}
	news, count, err := n.db.News(offset, n.newsPerPage, search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//Заполняем поле объекта паджинации
	p.Pages = count / n.newsPerPage
	if count%n.newsPerPage != 0 {
		p.Pages++
	}
	newsResponse.News = news
	newsResponse.Pagination = p

	// Возвращаем массив новостей с объектом паджинации
	bytes, err := json.Marshal(newsResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// ОБработчик получения детальной новости
func (n *newsService) detailedNewsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	news, err := n.db.NewsByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Возвращаем детальную новость
	bytes, err := json.Marshal(news)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
