# syntax=docker/dockerfile:1

FROM golang:1.23.2-alpine3.20 AS gobuilder
ADD . /src/
WORKDIR /src
RUN apk add --no-cache git && go mod download && go build -o /uyulala

FROM node:20-alpine3.20 AS nodebuilder
ADD ./frontend /src
WORKDIR /src/
RUN apk add --no-cache git file && npm i -g pnpm@latest 
RUN pnpm i
RUN pnpm run build

FROM alpine:3.20.0
COPY --from=gobuilder /uyulala /usr/bin/uyulala
COPY --from=nodebuilder /src/dist /www
COPY uyulala.docker.yml /etc/uyulala/uyulala.yml
CMD ["uyulala", "serve"]
