FROM golang:1.26.6-alpine@sha256:af8d6740070b8906d12eae1c3e3ea0957fb63f492051ea05e354c38ef9fe88df AS builder

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


FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b

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
