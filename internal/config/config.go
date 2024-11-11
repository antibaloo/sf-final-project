package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Структура конфигурации
type Config struct {
	postgresUser      string
	postgresPass      string
	postgresDatabase  string
	apiGatewayAddress string
	newsAddress       string
	commentsAddress   string
	censorAddress     string
	newsPerPage       int
	rssConfig         []byte
}

// Конструтктор структуры конфигурации
func NewConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return &Config{}, err
	}
	postgresUser, exist := os.LookupEnv("POSTGRES_USERNAME")
	if !exist {
		return &Config{}, errors.New("POSTGRES_USERNAME not found")
	}
	postgresPass, exist := os.LookupEnv("POSTGRES_PASSWORD")
	if !exist {
		return &Config{}, errors.New("POSTGRES_PASSWORD not found")
	}
	postgresDatabase, exist := os.LookupEnv("POSTGRES_DATABASE")
	if !exist {
		return &Config{}, errors.New("POSTGRES_DATABASE not found")
	}
	apiGatewayAddress, exist := os.LookupEnv("APIGATEWAY_ADDRESS")
	if !exist {
		return &Config{}, errors.New("APIGATEWAY_ADDRESS not found")
	}
	newsAddress, exist := os.LookupEnv("NEWS_ADDRESS")
	if !exist {
		return &Config{}, errors.New("NEWS_ADDRESS not found")
	}
	commentsAddress, exist := os.LookupEnv("COMMENTS_ADDRESS")
	if !exist {
		return &Config{}, errors.New("COMMENTS_ADDRESS not found")
	}
	censorAddress, exist := os.LookupEnv("CENSOR_ADDRESS")
	if !exist {
		return &Config{}, errors.New("CENSOR_ADDRESS not found")
	}

	newsPerPageStr, exist := os.LookupEnv("NEWS_PER_PAGE")
	if !exist {
		return &Config{}, errors.New("NEWS_PER_PAGE not found")
	}
	newsPerPage, err := strconv.Atoi(newsPerPageStr)
	if err != nil {
		return &Config{}, fmt.Errorf("NEWS_PER_PAGE conversion error: %s", err.Error())
	}
	if newsPerPage < 1 {
		return &Config{}, fmt.Errorf("NEWS_PER_PAGE need to bo over 1")
	}
	rssConfigFile, exist := os.LookupEnv("RSS_CONFIG")
	if !exist {
		return &Config{}, errors.New("RSS_CONFIG not found")
	}
	// чтение файла конфигурации rss
	rssConfig, err := os.ReadFile(rssConfigFile)
	if err != nil {
		return &Config{}, fmt.Errorf("error while reading rss configuration file: %s", err.Error())
	}
	return &Config{
		postgresUser,
		postgresPass,
		postgresDatabase,
		apiGatewayAddress,
		newsAddress,
		commentsAddress,
		censorAddress,
		newsPerPage,
		rssConfig,
	}, nil
}

func (c *Config) ConString() string {
	return fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable", c.postgresUser, c.postgresPass, c.postgresDatabase)
}

func (c *Config) APIGatewayAddress() string {
	return c.apiGatewayAddress
}

func (c *Config) NewsAddress() string {
	return c.newsAddress
}

func (c *Config) CommentsAddress() string {
	return c.commentsAddress
}

func (c *Config) CensorAddress() string {
	return c.censorAddress
}

func (c *Config) NewsPerPage() int {
	return c.newsPerPage
}

func (c *Config) RssConfig() []byte {
	return c.rssConfig
}
