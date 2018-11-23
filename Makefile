-include .env

PROJECT_NAME := vote-bot
CONFIG_FILE := config.json

MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
CURRENT_DIR := $(patsubst %/,%,$(dir $(MAKEFILE_PATH)))

all: container run

build:
	@go build -v

container:
	@docker build --build-arg aws_access_key_id=${AWS_ACCESS_KEY_ID} --build-arg aws_secret_access_key=${AWS_SECRET_ACCESS_KEY} -t ${PROJECT_NAME} .

run:
	@docker run -d -v ${CURRENT_DIR}/${CONFIG_FILE}:/app/config.json --name ${PROJECT_NAME} ${PROJECT_NAME}