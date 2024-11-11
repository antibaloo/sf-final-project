-- +goose Up
-- +goose StatementBegin
CREATE TABLE dictionary(
    word TEXT NOT NULL UNIQUE
);

INSERT INTO dictionary (word) VALUES ('xxx'),('yyy'),('zzz');

CREATE TABLE news(
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL CHECK(title <>''),
    content TEXT NOT NULL CHECK(title <>''),
    pub_time INT DEFAULT 0,
    link TEXT UNIQUE NOT NULL CHECK(title <>'')
);

CREATE TABLE comments(
    id SERIAL PRIMARY KEY,
    news_id INT NOT NULL,
    comment_id INT,
    content TEXT NOT NULL CHECK(content <>''),
    created_at INT NOT NULL CHECK(created_at <> 0),
    updated_at INT NOT NULL CHECK(updated_at <> 0),
    CONSTRAINT fk_comments_news_id
        FOREIGN KEY (news_id)
            REFERENCES news (id)
            ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS news;
DROP TABLE IF EXISTS dictionary;
-- +goose StatementEnd
