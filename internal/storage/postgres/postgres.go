package postgres

import (
	"context"
	"time"

	"github.com/antibaloo/sf-final-project/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Структура хоанилища PosgreSQL
type Store struct {
	Pool *pgxpool.Pool
}

// Конструктор хранилища
func NewStore(conStr string) (*Store, error) {
	db, err := pgxpool.New(context.Background(), conStr)
	if err != nil {
		return &Store{}, err
	}
	return &Store{Pool: db}, nil
}

// Метод добавления новости
func (s *Store) AddNews(news storage.NewsShortDetailed) error {
	_, err := s.Pool.Exec(
		context.Background(),
		"INSERT INTO news(title, content, pub_time, link) VALUES ($1, $2, $3, $4)",
		news.Title,
		news.Content,
		news.PubTime,
		news.Link,
	)
	if err != nil {
		return err
	}
	return nil
}

// Метод получения списка новостей
func (s *Store) News(offset, limit int, search string) ([]storage.NewsShortDetailed, int, error) {
	var (
		news  []storage.NewsShortDetailed
		count int
	)
	// Получаем общее число строк в ответе
	err := s.Pool.QueryRow(
		context.Background(),
		`SELECT count(*) FROM news WHERE LOWER(title) LIKE '%`+search+`%'`,
	).Scan(&count)
	if err != nil {
		return news, 0, err
	}

	// Получаем только строки с нужной страницы
	rows, err := s.Pool.Query(
		context.Background(),
		`SELECT id, title, content, pub_time, link FROM news WHERE LOWER(title) LIKE '%`+search+`%' ORDER BY id DESC OFFSET $1 LIMIT $2`,
		offset,
		limit,
	)
	if err != nil {
		return news, 0, err
	}
	// Итерируем по строкам, записываем результат
	for rows.Next() {
		n := storage.NewsShortDetailed{}
		err := rows.Scan(
			&n.Id,
			&n.Title,
			&n.Content,
			&n.PubTime,
			&n.Link,
		)
		if err != nil {
			return news, 0, err
		}
		news = append(news, n)
	}
	if rows.Err() != nil {
		return news, 0, rows.Err()
	}
	return news, count, nil
}

// Метод получения детальной новости по идентификатору
func (s *Store) NewsByID(id int) (storage.NewsShortDetailed, error) {
	var news storage.NewsShortDetailed

	err := s.Pool.QueryRow(
		context.Background(),
		`SELECT id, title, content, pub_time, link FROM news WHERE id = $1`,
		id,
	).Scan(
		&news.Id,
		&news.Title,
		&news.Content,
		&news.PubTime,
		&news.Link,
	)
	if err != nil {
		return storage.NewsShortDetailed{}, err
	}
	return news, nil
}

// Метод добавления коментария
func (s *Store) AddComment(comment storage.Comment) error {
	_, err := s.Pool.Exec(
		context.Background(),
		`INSERT INTO comments (news_id, comment_id, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		comment.NewsId,
		comment.CommentId,
		comment.Content,
		time.Now().Unix(),
		time.Now().Unix(),
	)
	if err != nil {
		return err
	}
	return nil
}

// Метод получение списка комментариев к новости
func (s *Store) CommentsByNewsId(id int) ([]storage.Comment, error) {
	var comments []storage.Comment

	rows, err := s.Pool.Query(
		context.Background(),
		`SELECT id, news_id, comment_id, content, created_at, updated_at FROM comments WHERE news_id = $1`,
		id,
	)
	if err != nil {
		return comments, err
	}
	for rows.Next() {
		var comment storage.Comment
		err := rows.Scan(
			&comment.Id,
			&comment.NewsId,
			&comment.CommentId,
			&comment.Content,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return comments, err
		}
		comments = append(comments, comment)
	}
	if rows.Err() != nil {
		return comments, rows.Err()
	}
	return comments, nil
}

// Метод получения словаря запрещенных слов
func (s *Store) Dictionary() ([]string, error) {
	var words []string
	rows, err := s.Pool.Query(
		context.Background(),
		`SELECT word FROM dictionary`,
	)
	if err != nil {
		return []string{}, err
	}
	for rows.Next() {
		var word string
		err = rows.Scan(
			&word,
		)
		if err != nil {
			return []string{}, err
		}
		words = append(words, word)
	}
	return words, nil
}

// Метод добавления запрещенного слова в словарь
func (s *Store) AddWord2Dictionary(word string) error {
	_, err := s.Pool.Exec(
		context.Background(),
		`INSERT INTO dictionary (word) VALUES ($1)`,
		word,
	)
	if err != nil {
		return err
	}
	return nil
}
