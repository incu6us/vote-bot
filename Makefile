-include .env

PROJECT_NAME := vote-bot
CONFIG_PATH := "."


all: container run

build:
	@go mod tidy && go build -v

container:
	@docker build -t ${PROJECT_NAME} .

run:
	@docker run -d -v ${CONFIG_PATH}/config.json:/app/config.json --build-arg AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} --build-arg AWS_ACCESS_KEY=${AWS_ACCESS_KEY} --name vote-bot vote-bot