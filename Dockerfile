FROM golang:1.22.6-alpine AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN apk add build-base && go mod download
COPY . .
RUN CGO_ENABLED=1 go build -ldflags="-s -w -X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o /fly2stats

FROM alpine as app
COPY --from=build /fly2stats /fly2stats
ENTRYPOINT [ "/fly2stats" ]