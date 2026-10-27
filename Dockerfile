# syntax=docker/dockerfile:1

# Bản Go dùng để BUILD. go.mod chỉ ghi bản tối thiểu; build luôn dùng bản còn nhận vá bảo mật (https://go.dev/dl).
ARG GO_VERSION=1.27.3

# ===== build =====
# --platform=$BUILDPLATFORM: biên dịch chéo ngay trên máy build, không giả lập CPU khi làm image multi-arch.
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

# Dependency trước, code sau: đổi code không phải tải lại module.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/app .

# ===== runtime =====
# distroless/static: không shell, không package manager; có sẵn CA certs + tzdata.
# Cần build tái lập được thì pin thêm digest: ...static-debian12:nonroot@sha256:<digest>
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
# UID dạng SỐ để Kubernetes runAsNonRoot kiểm được (user dạng tên thì kubelet từ chối chạy).
USER 65532:65532
EXPOSE 8080
# Exec form: app nhận thẳng SIGTERM; tắt êm đã xử lý trong main.go + framework/service/httpx.
ENTRYPOINT ["/app"]
