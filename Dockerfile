# Build nfs-ganesha-metrics
FROM golang:1.19 as builder
ARG GIT_VERSION="(unset)"
ARG COMMIT_ID="(unset)"
ARG ARCH=""

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd cmd
COPY internal internal

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${ARCH} GO111MODULE=on \
    go build -a \
    -ldflags "-X main.Version=${GIT_VERSION} -X main.CommitID=${COMMIT_ID}" \
    -o nfsgmetrics cmd/main.go

# Create stand-alone go image
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest
COPY --from=builder /workspace/nfsgmetrics /bin/nfsgmetrics

ENTRYPOINT ["/bin/nfsgmetrics"]
EXPOSE 8080
