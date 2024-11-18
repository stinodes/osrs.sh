FROM golang:alpine
# RUN apk update && apk add --no-cache git && apk add --no-cache bash && apk add build-base

ENV TERM=xterm-256color
ENV COLORTERM=truecolor
ENV CI=1
ENV CLICOLOR_FORCE=1


WORKDIR /usr/app

RUN go install github.com/air-verse/air@latest

COPY . .
RUN go mod tidy

CMD ["air", "./src/cmd/rest.go", "-b", "0.0.0.0"]
