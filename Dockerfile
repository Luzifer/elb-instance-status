FROM golang:1.26-alpine@sha256:f85330846cde1e57ca9ec309382da3b8e6ae3ab943d2739500e08c86393a21b1 AS builder

COPY . /src/elb-instance-status
WORKDIR /src/elb-instance-status

ENV SOURCE_DATE_EPOCH=1
ENV LDFLAGS="-w -s -buildid="
ENV GO_LDFLAGS=${LDFLAGS}

RUN set -ex \
 && apk add --update git \
 && go install \
		  -buildvcs=false \
      -ldflags "-X main.version=$(git describe --tags --always || echo dev)" \
      -mod=readonly \
      -modcacherw \
      -trimpath


FROM alpine:3.23

LABEL org.opencontainers.image.authors="Knut Ahlers <knut@ahlers.me>" \
      org.opencontainers.image.url="https://github.com/Luzifer/elb-instance-status/pkgs/container/elb-instance-status" \
      org.opencontainers.image.source="https://github.com/Luzifer/elb-instance-status" \
      org.opencontainers.image.title="elb-instance-status"

RUN set -ex \
 && apk --no-cache add \
      bash \
      ca-certificates \
      curl \
      jq \
      yq-go

COPY --from=builder /go/bin/elb-instance-status /usr/bin/elb-instance-status

ENTRYPOINT ["/usr/bin/elb-instance-status"]
