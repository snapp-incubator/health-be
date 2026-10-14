#build stage
FROM golang:1.27.2 AS builder
WORKDIR /go/src/app

COPY go.mod /go/src/app/
RUN go mod download

COPY . /go/src/app
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o main

#final stage
FROM gcr.io/distroless/static-debian13
WORKDIR /app
COPY --from=builder /go/src/app/main /app/main
CMD ["/app/main"]
