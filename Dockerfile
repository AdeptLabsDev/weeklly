# syntax=docker/dockerfile:1.7

# ---------------------------------------------------------------------------
# Estágio 1: build. Debian, não Alpine: o CLI standalone do Tailwind é um
# binário glibc. tools/tailwind escolhe a variante certa por GOOS/GOARCH e
# confere o checksum antes de usar.
# ---------------------------------------------------------------------------
FROM golang:1.27-bookworm AS build
WORKDIR /src

# Dependências antes do código: esta camada só invalida quando go.mod/go.sum mudam.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go run ./tools/tailwind \
 && ./.tools/tailwindcss -i web/styles/app.css -o web/static/app.css --minify \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/weeklly ./cmd/weeklly \
 && mkdir -p /out/data

# ---------------------------------------------------------------------------
# Estágio 2: runtime. Sem shell, sem gerenciador de pacotes, sem root.
# static-debian12 traz só CA certs e tzdata; o binário é estático (CGO_ENABLED=0).
# ---------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/weeklly /weeklly
# /data precisa existir na imagem com dono nonroot (uid 65532): um volume
# nomeado herda o dono do diretório da imagem no primeiro mount.
COPY --from=build --chown=65532:65532 /out/data /data

ENV WEEKLLY_ENV=production \
    WEEKLLY_ADDR=:8080 \
    WEEKLLY_DB_PATH=/data/weeklly.db

EXPOSE 8080
USER 65532:65532

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 CMD ["/weeklly", "-healthcheck"]
ENTRYPOINT ["/weeklly"]
