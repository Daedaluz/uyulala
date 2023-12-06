FROM golang:1.21-alpine3.18 AS gobuilder
ADD . /src/
WORKDIR /src
RUN apk add --no-cache git && go build -o /uyulala

FROM node:18-alpine3.18 AS nodebuilder
ADD . /src/
WORKDIR /src/frontend
RUN apk add --no-cache git && npm i && npm run build

FROM alpine:3.18
COPY --from=gobuilder /uyulala /usr/bin/uyulala
COPY --from=nodebuilder /src/frontend/dist /www
COPY uyulala.docker.yml /etc/uyulala.yml
CMD ["uyulala", "serve"]
