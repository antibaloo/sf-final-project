package middleware

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Структура измененного http.ResponseWriter для извлечения кода ответа
type customRW struct {
	http.ResponseWriter
	statusCode int
	statusText string
}

// Конструктор для кастомного RW
func newCRW(w http.ResponseWriter) *customRW {
	return &customRW{w, http.StatusOK, http.StatusText(http.StatusOK)}
}

// Возвращает зписанный StatusCode
func (crw *customRW) StatusCode() int {
	return crw.statusCode
}

// Возвращает зписанный StatusText
func (crw *customRW) StatusText() string {
	return crw.statusText
}

// Перегруженная функция записи заголовка
func (crw *customRW) WriteHeader(code int) {
	crw.statusCode = code
	crw.statusText = http.StatusText(code)
	crw.ResponseWriter.WriteHeader(code)
}

// Метод для генерации идентификатора запроса и логирования
func GenIdAndLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Генерим уникальный идентификатор запроса, если его нет в исходном запросе
		request_id := r.URL.Query().Get("request_id")
		if request_id == "" {
			generator := rand.New(rand.NewSource(time.Now().UnixNano()))
			request_id = strconv.FormatInt(generator.Int63(), 10)
			//Добавляем сгенеренный идентификатор в r.URL.Query
			q := r.URL.Query()
			q.Add("request_id", request_id)
			r.URL.RawQuery = q.Encode()
		}
		// Создаем кастомный responseWriter, чтобы иметь возможность читать код ответа из следующего обработчика
		crw := newCRW(w)
		next.ServeHTTP(crw, r)
		// Логируем запрос и результат его работы
		fmt.Printf("%v: от %v запрос по адресу: %v, ответ: %v - %s, идентификатор: %s\n", time.Now().Format("02.01.2006 15:04:05 MST"), r.RemoteAddr, r.RequestURI, crw.StatusCode(), crw.StatusText(), request_id)
	})
}
