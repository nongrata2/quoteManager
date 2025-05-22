FROM golang:1.23 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/quotemanager ./cmd/quotemanager
COPY internal ./internal
COPY migrations ./migrations
COPY pkg ./pkg

RUN CGO_ENABLED=0 go build -o /quotemanager ./cmd/quotemanager/main.go

FROM alpine:3.20

COPY --from=build /quotemanager /quotemanager

ENTRYPOINT ["/quotemanager"]