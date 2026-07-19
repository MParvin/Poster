# Override images when Docker Hub is unreachable, e.g.:
#   --build-arg GO_IMAGE=focker.ir/golang:1.26.2-alpine3.22
ARG GO_IMAGE=golang:1.26-alpine
ARG RUNTIME_IMAGE=alpine:3.24

# ---- Builder Stage ----
# hadolint ignore=DL3006
FROM ${GO_IMAGE} AS builder

ARG ALPINE_MIRROR=
ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}
ENV GOTOOLCHAIN=local

# hadolint ignore=DL3018
RUN set -eux; \
    if [ -n "${ALPINE_MIRROR}" ]; then \
      ver="v$(cut -d. -f1-2 /etc/alpine-release)"; \
      { \
        echo "${ALPINE_MIRROR}/${ver}/main"; \
        echo "${ALPINE_MIRROR}/${ver}/community"; \
      } > /etc/apk/repositories; \
    fi; \
    for attempt in 1 2 3; do \
      apk update && apk add --no-cache git ca-certificates && break; \
      [ "${attempt}" -eq 3 ] && exit 1; \
      sleep "${attempt}"; \
    done

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /social_poster .

# ---- Final Stage ----
# hadolint ignore=DL3006
FROM ${RUNTIME_IMAGE}

ARG ALPINE_MIRROR=

# hadolint ignore=DL3018
RUN set -eux; \
    if [ -n "${ALPINE_MIRROR}" ]; then \
      ver="v$(cut -d. -f1-2 /etc/alpine-release)"; \
      { \
        echo "${ALPINE_MIRROR}/${ver}/main"; \
        echo "${ALPINE_MIRROR}/${ver}/community"; \
      } > /etc/apk/repositories; \
    fi; \
    for attempt in 1 2 3; do \
      apk update && apk add --no-cache git ca-certificates su-exec && break; \
      [ "${attempt}" -eq 3 ] && exit 1; \
      sleep "${attempt}"; \
    done; \
    addgroup -S poster && adduser -S poster -G poster -u 1000 -h /app

WORKDIR /app

COPY --from=builder /social_poster /app/social_poster
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod 755 /docker-entrypoint.sh /app/social_poster && chown poster:poster /app/social_poster

ENV POSTS_REPO_PATH=/app/my_posts_data \
    HOME=/app \
    TMPDIR=/tmp \
    PROCESS_INTERVAL=5m

HEALTHCHECK --interval=60s --timeout=5s --start-period=30s --retries=3 \
  CMD ["/app/social_poster", "health"]

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/app/social_poster", "process"]
