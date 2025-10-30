FROM golang:1.25-alpine AS builder
WORKDIR /src
WORKDIR /src/FGoFile
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/fgofile .

FROM alpine:3.20
RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /out/fgofile /usr/local/bin/fgofile
USER app
EXPOSE 2121
ENTRYPOINT ["/usr/local/bin/fgofile"]
