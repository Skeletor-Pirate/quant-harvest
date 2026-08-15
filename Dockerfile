FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /qharvest .

# Create the data directory and set ownership for the distroless nonroot user (UID 65532)
RUN mkdir -p /app/data && chown -R 65532:65532 /app/data

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /qharvest /qharvest
# Copy the pre-created data directory with correct nonroot ownership
COPY --from=build --chown=nonroot:nonroot /app/data /app/data
VOLUME ["/app/data"]
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/qharvest", "-store", "/app/data/qharvest.db"]
