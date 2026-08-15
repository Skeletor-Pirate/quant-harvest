FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /qharvest .

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /qharvest /qharvest
VOLUME ["/app/data"]
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/qharvest", "-store", "/app/data/qharvest.db"]
