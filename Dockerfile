FROM golang:1.15.7-alpine3.13
RUN mkdir /app
ADD main.go /app
ADD model.go /app
ADD app.go /app
ADD go.mod /app
WORKDIR /app
RUN go mod download
RUN go build -o main .
CMD ["/app/main"]