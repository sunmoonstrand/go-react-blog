FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git gcc musl-dev

# 初始化新模块
COPY . .
RUN if [ ! -f go.mod ]; then go mod init blog; fi

CMD ["go", "run", "main.go"]