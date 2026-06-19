FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o /karots-drop ./cmd/karots-drop/

FROM scratch
COPY --from=builder /karots-drop /karots-drop
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s CMD ["/karots-drop", "health"]
ENTRYPOINT ["/karots-drop"]
CMD ["serve"]
