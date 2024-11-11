package rss

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/antibaloo/sf-final-project/internal/storage"
)

var errDuplicate = `ERROR: duplicate key value violates unique constraint "news_link_key" (SQLSTATE 23505)`

// Набор вложенных структур для раскодировки xml rss фида
type feed struct {
	RSS     string  `xml:"rss"`
	Channel channel `xml:"channel"`
}

type channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Items       []item `xml:"item"`
}

type item struct {
	Title   string `xml:"title"`
	Content string `xml:"description"`
	Link    string `xml:"link"`
	PubTime string `xml:"pubDate"`
}

type rssReader struct {
	URLs          []string      `json:"rss"`
	RequestPeriod time.Duration `json:"request_period"`
	db            storage.Store
}

// Метод создает структуру ридера новостей
func CreateService(config []byte, db storage.Store) (*rssReader, error) {
	var rss rssReader
	err := json.Unmarshal(config, &rss)
	if err != nil {
		return &rssReader{}, err
	}
	rss.db = db
	return &rss, nil
}

// Метод запускает ридер новостей по одному на каждый rss канал
func (r *rssReader) Start() {
	for _, url := range r.URLs {
		go readNews(r.db, url, r.RequestPeriod)
	}
}

// Метод читает новости из канала с заданный периодом
func readNews(db storage.Store, url string, period time.Duration) {
	fmt.Printf("%v: чтение новостей из канала %s начато\n", time.Now().Format("02.01.2006 15:04:05 MST"), url)
	for {
		news, err := parseFeed(url)
		if err != nil {
			fmt.Printf(
				"%v: при чтении новостей из канала %s произошла ошибка: %s\n",
				time.Now().Format("02.01.2006 15:04:05 MST"),
				url, err.Error(),
			)
			continue
		}
		countNews := 0
		for _, n := range news {
			err := db.AddNews(n)
			if err != nil {
				// Игнорируем ошибку дубликата уникального поля, т.к. сами его сделали (поле ссылка на новость уникально для
				// предотвращения повторно запсии новости в БД)
				if err.Error() != errDuplicate {
					fmt.Printf(
						"%v: при попытке записи новости из канала %s в БД произошла ошибка: %s\n",
						time.Now().Format("02.01.2006 15:04:05 MST"),
						url, err.Error(),
					)
				}
			} else {
				countNews++
			}
		}
		fmt.Printf("%v: получено %d новостей из фида: %s \n", time.Now().Format("02.01.2006 15:04:05 MST"), countNews, url)
		time.Sleep(time.Minute * period)
	}
}

// Метод разбирает новости из канала
func parseFeed(url string) ([]storage.NewsShortDetailed, error) {
	// Сохраняем ответ на запрос по адресу url
	response, err := http.Get(url)
	if err != nil {
		return []storage.NewsShortDetailed{}, err
	}

	// Читаем тело ответа в массив байт
	b, err := io.ReadAll(response.Body)
	if err != nil {
		return []storage.NewsShortDetailed{}, err
	}

	var feed feed
	// Раскодиреум xml в структуру
	err = xml.Unmarshal(b, &feed)
	if err != nil {
		return []storage.NewsShortDetailed{}, err
	}

	var news []storage.NewsShortDetailed
	// Итерируем по массиву новостей
	for _, item := range feed.Channel.Items {
		var n storage.NewsShortDetailed
		n.Title = item.Title
		// Удаляем html тэги с помощью регулярного выражения
		n.Content = stripHtmlTags(item.Content)
		n.Link = item.Link
		// Парсим время публикации по одному формату
		t, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700", item.PubTime)
		// Если получаем ошибку, то парсим другой формат
		if err != nil {
			t, err = time.Parse("Mon, 2 Jan 2006 15:04:05 GMT", item.PubTime)
		}
		if err != nil {
			return []storage.NewsShortDetailed{}, err
		}
		n.PubTime = t.Unix()
		news = append(news, n)
	}
	return news, nil
}

func stripHtmlTags(s string) string {
	// Регулярное выражение для удаления html тэгов
	const regex = `<.*?>`
	r := regexp.MustCompile(regex)
	// Удаляем html тэги с помощью регулярного выражения
	return r.ReplaceAllString(s, "")
}
