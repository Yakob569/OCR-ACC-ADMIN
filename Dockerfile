FROM golang:1.26-alpine AS build

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo -ldflags="-s -w" -o /out/app ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=build /out/app /app/app
COPY db/migrations /app/db/migrations

# Render injects PORT at runtime; default matches local admin dev.
ENV PORT=8082
EXPOSE 8082

CMD ["/app/app"]
