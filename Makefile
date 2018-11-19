-include .env

PROJECT_NAME := vote-bot
CONFIG_PATH := "."


all: container run

build:
	@go mod tidy && go build -v

container:
	@docker build --build-arg AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} --build-arg AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} -t ${PROJECT_NAME} .

run:
	@docker run -d -v ${CONFIG_PATH}/config.json:/app/config.json --name ${PROJECT_NAME} ${PROJECT_NAME}