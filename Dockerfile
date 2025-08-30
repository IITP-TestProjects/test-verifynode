ARG GO_VER=1.25
ARG ALPINE_VER=3.21
ARG PORT

FROM alpine:${ALPINE_VER} AS peer-base
RUN apk add --no-cache tzdata

RUN echo 'hosts: files dns' > /etc/nsswitch.conf

FROM golang:${GO_VER}-alpine${ALPINE_VER} AS golang
RUN apk add --no-cache \
	bash \
	gcc \
	git \
	make \
	musl-dev
ADD . $GOPATH/src/test_verifier
WORKDIR $GOPATH/src/test_verifier

FROM golang AS peer
RUN go build -o verifier

FROM peer-base
COPY --from=peer /go/src/test_verifier /usr/local/bin
EXPOSE 50053
CMD ["verifier"]