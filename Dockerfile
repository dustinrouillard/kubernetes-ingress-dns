FROM golang:1.18-alpine AS builder

ENV GO111MODULE=on
ENV CGO_ENABLED=0

WORKDIR /go/src/github.com/dustinrouillard/kubernetes-ingress-dns

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build
RUN chmod +x /go/src/github.com/dustinrouillard/kubernetes-ingress-dns/kubernetes-ingress-dns

FROM alpine

RUN apk update
RUN apk add dnsmasq bind-tools
RUN mkdir /hosts

ADD dnsmasq/dnsmasq.conf /etc/dnsmasq/dnsmasq.conf
ADD dnsmasq/docker-entrypoint.sh /

COPY --from=builder /go/src/github.com/dustinrouillard/kubernetes-ingress-dns/kubernetes-ingress-dns /bin/kubernetes-ingress-dns

ENTRYPOINT ["/docker-entrypoint.sh"]