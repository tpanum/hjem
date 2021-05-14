FROM golang:1.16 as builder
RUN mkdir /build
WORKDIR /build
COPY *.go go.mod go.sum ./
COPY app app
COPY frontend frontend
RUN ls -lha /build
WORKDIR /build/app
RUN CGO_ENABLED=1 GOOS=linux go build -a -tags netgo -ldflags "-linkmode external -extldflags -static" -o hjem main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
RUN mkdir app data
WORKDIR /app
COPY --from=builder /build/app/hjem /app

ENTRYPOINT [ "./hjem"]

CMD [ "-db-file" "/data/hjem.db" ]
