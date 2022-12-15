FROM golang:1.19-alpine AS builder
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64
RUN apk update && apk add make
RUN mkdir /app
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make build

FROM scratch
COPY --from=builder /app/templates/ /templates/
COPY --from=builder /app/static/ /static/
COPY --from=builder /app/out/bin /
CMD [ "/nacre-server" ]