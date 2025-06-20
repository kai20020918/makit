# ビルドステージ
FROM golang:1-bullseye AS builder
WORKDIR /work
ARG CGO_ENABLED=0 [cite: 20]
COPY . .
RUN go build -o makit cmd/main/makit.go

# 配布用イメージステージ
FROM scratch
ARG VERSION=0.5.1
LABEL org.opencontainers.image.source=https://github.com/kai20020918/makit \
      org.opencontainers.image.version=${VERSION} \
      org.opencontainers.image.title=makit \
      org.opencontainers.image.description="Another implementation of Word Count (wc), again."

# RUN adduser -disabled-password --disabled-login --home /workdir nonroot \
#     && mkdir -p /workdir  # nonrootユーザーを追加
COPY --from=builder /work/makit /opt/wildcherry/makit
COPY --from=golang:1.12 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /workdir
# USER nonroot [cite: 20] # ユーザーをnonrootに変更
ENTRYPOINT [ "/opt/wildcherry/makit" ]