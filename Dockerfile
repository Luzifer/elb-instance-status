FROM golang:1.27.3-alpine@sha256:b8c9b82affbbb82a020bbcf998dad2ba9516944ac7e883524642a481a9505e9c AS builder

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


FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a

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
