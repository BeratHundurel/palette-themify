FROM golang:1.26.2-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/themesmith-api .

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S themesmith \
    && adduser -S -G themesmith themesmith
COPY --from=build /out/themesmith-api /usr/local/bin/themesmith-api
USER themesmith
EXPOSE 8088
ENTRYPOINT ["themesmith-api"]
