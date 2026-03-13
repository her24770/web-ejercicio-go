FROM golang:1.22-alpine

WORKDIR /app

EXPOSE 24770

CMD ["go", "run", "main.go"]