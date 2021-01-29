FROM golang:1.15.7-alpine3.13
RUN mkdir /app
ADD main.go /app
ADD booking.go /app
ADD bookingConfig.go /app
ADD facilityDetail.go /app
ADD account.go /app
ADD app.go /app
ADD go.mod /app
WORKDIR /app
RUN go mod download
RUN go build -o main .
CMD ["/app/main"]