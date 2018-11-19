### Build
FROM golang:1.11 AS build

ENV APP_ROOT_DIR=/build
WORKDIR ${APP_ROOT_DIR}
COPY . .

RUN make build

### Main
FROM alpine:3.8

ARG AWS_ACCESS_KEY_ID
ARG AWS_ACCESS_KEY

ENV APP_ROOT_DIR=/app
RUN mkdir ${APP_ROOT_DIR}
WORKDIR ${APP_ROOT_DIR}

RUN apk add --no-cache ca-certificates openssl libstdc++ libc6-compat

COPY --from=build /build/vote-bot ${APP_ROOT_DIR}/

RUN echo -e '#!/bin/sh\n${APP_ROOT_DIR}/vote-bot' > startup.sh
RUN chmod +x vote-bot startup.sh

RUN pwd
RUN ls -la

ENTRYPOINT ["./startup.sh"]