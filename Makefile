.PHONY: build-apigateway
build-apigateway:
	@go build -o build/apigateway -v ./cmd/apigateway

.PHONY: build-news
build-news:
	@go build -o build/news -v ./cmd/news

.PHONY: build-comments
build-comments:
	@go build -o build/comments -v ./cmd/comments

.PHONY: build-censor
build-censor:
	@go build -o build/censor -v ./cmd/censor

.PHONY: build
build:
	@go build -o build/apigateway -v ./cmd/apigateway
	@go build -o build/news -v ./cmd/news
	@go build -o build/comments -v ./cmd/comments
	@go build -o build/censor -v ./cmd/censor

.PHONY: run-apigateway
run-apigateway:
	@go run cmd/apigateway/apigateway.go && disown apigateway

.PHONY: run-news
run-news:
	@go run cmd/news/news.go

.PHONY: run-comments
run-comments:
	@go run cmd/comments/comments.go
	
.PHONY: run-censor
run-censor:
	@go run cmd/censor/censor.go
