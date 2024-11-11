package storage

// Структура комментария
type Comment struct {
	Id        int    `json:"id"`         // Идентификатор комментария, первичный ключ
	NewsId    int    `json:"news_id"`    // Идентификатор новости, к которой дан комментарий
	CommentId int    `json:"comment_id"` // Идентификатор комментария, ответом к которому выступает комментарий
	Content   string `json:"content"`    // Сам комментарий
	CreatedAt int64  `json:"created_at"` // Время создания комментария
	UpdatedAt int64  `json:"updated_at"` // Время последнего изменения в комментарии
}

// Структура сокращенной новости
type NewsShortDetailed struct {
	Id      int    `json:"id"`       //Идентификатор
	Title   string `json:"title"`    //Заголовок новости
	Content string `json:"content"`  // Первый абзац новости
	PubTime int64  `json:"pub_time"` // Время публикации новости в источнике
	Link    string `json:"link"`     // Ссылка на источник
}

// Структура детальной новости
type NewsFullDetailed struct {
	NewsShortDetailed
	Comments []Comment `json:"comments"` // Комментарии к новости
}

// Контракт на методы  хранилища
type Store interface {
	AddNews(NewsShortDetailed) error
	News(int, int, string) ([]NewsShortDetailed, int, error)
	NewsByID(int) (NewsShortDetailed, error)
	AddComment(Comment) error
	CommentsByNewsId(int) ([]Comment, error)
	Dictionary() ([]string, error)
	AddWord2Dictionary(string) error
}
