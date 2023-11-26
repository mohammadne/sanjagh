FROM golang:alpine AS builder
RUN apk add git
WORKDIR /sanjagh
COPY . .
RUN go build -o /usr/local/bin/sanjagh

FROM alpine:latest AS runtime
LABEL maintainer="Mohammad Nasr <mohammadne.dev@gmail.com>"
COPY --from=builder /usr/local/bin/sanjagh /usr/local/bin/sanjagh
ENTRYPOINT ["/usr/local/bin/sanjagh"]
