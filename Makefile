-include .env

PROJECT_NAME := vote-bot


all: build

build:
	@go mod tidy && go build -v

container:
	@docker build -t ${PROJECT_NAME} .