# syntax=docker/dockerfile:1

FROM golang:1.22-alpine3.19 AS gobuilder
ADD . /src/
WORKDIR /src
RUN apk add --no-cache git && go mod download && go build -o /uyulala

FROM node:20-alpine3.19 AS nodebuilder
ADD ./frontend /src
WORKDIR /src/
RUN apk add --no-cache git file && npm i -g pnpm@latest 
RUN npm i
RUN npm run build || (cat /root/.npm/_logs/* && env && file /bin/sh && exit 1)

FROM alpine:3.19
COPY --from=gobuilder /uyulala /usr/bin/uyulala
COPY --from=nodebuilder /src/dist /www
COPY uyulala.docker.yml /etc/uyulala/uyulala.yml
CMD ["uyulala", "serve"]
