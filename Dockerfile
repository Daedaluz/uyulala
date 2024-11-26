# syntax=docker/dockerfile:1
FROM golang:1.23.3 as gobuilder
ADD . /src/
WORKDIR /src/
ARG CGO_ENABLED=0
RUN go build -o /uyulala

FROM node:20-alpine3.20 AS nodebuilder
ADD ./frontend /src
WORKDIR /src/
RUN apk add --no-cache git file && npm i -g pnpm@latest 
RUN pnpm i
RUN pnpm run build

FROM busybox:latest
COPY --from=gobuilder /uyulala /usr/bin/uyulala
COPY --from=nodebuilder /src/dist /www
COPY uyulala.docker.yml /etc/uyulala/uyulala.yml
CMD ["uyulala", "serve"]
