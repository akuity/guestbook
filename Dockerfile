FROM --platform=$BUILDPLATFORM golang:1.22 as builder
ARG TARGETOS TARGETARCH

WORKDIR /src
COPY go.mod /src/go.mod
COPY go.sum /src/go.sum
COPY . .
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make build

FROM alpine:latest
WORKDIR /app
COPY --from=builder /src/build/guestbook /app/guestbook
COPY ./public/index.html public/index.html
COPY ./public/script.js public/script.js
COPY ./public/style.css public/style.css
CMD ["/app/guestbook"]
EXPOSE 3000
USER 1000:0
