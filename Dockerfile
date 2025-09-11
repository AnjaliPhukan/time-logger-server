FROM golang:1.24-alpine AS builder

RUN mkdir /app
RUN mkdir /app/certs

COPY go.mod /app
COPY server.go /app

WORKDIR /app
RUN go build -o server .

FROM gcr.io/distroless/base AS runner

COPY certs/server.key /
COPY certs/server.crt /
COPY --from=builder /app/server /

CMD [ "/server", "-cert=server.crt", "-key=server.key" ]
EXPOSE 8443
