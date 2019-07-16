FROM golang:1.12.7-alpine AS builder

WORKDIR /go/src/app

COPY ./Gopkg.lock ./Gopkg.toml ./

RUN apk add --no-cache curl git mercurial \
    && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh \
    && dep ensure -v -vendor-only \
    && apk del curl git mercurial

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix gco -o ./bin/contacts-api ./src/main.go


FROM alpine:3.10
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/app/bin/contacts-api .
CMD ["./contacts-api"]
