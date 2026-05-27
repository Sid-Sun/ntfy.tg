FROM golang:1.22-bullseye as build

WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN make build

FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /root/app
COPY --from=build /build/bin/ntfy.tg .

EXPOSE 9090

CMD [ "/root/app/ntfy.tg" ]
