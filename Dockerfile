FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG RELEASE_VERSION=development

# Install our build tools
RUN apk add --update git make bash ca-certificates

WORKDIR /go/src/github.com/daspawnw/k8s-node-label
COPY . ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-X 'main.Version=${RELEASE_VERSION}'" -o bin/k8s-node-label ./cmd/k8s-node-label/...

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/daspawnw/k8s-node-label/bin/k8s-node-label /k8s-node-label

ENTRYPOINT ["/k8s-node-label"]