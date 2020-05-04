FROM golang:1.13-alpine as builder

ARG RELEASE_VERSION=development

# Install our build tools
RUN apk add --update git make bash ca-certificates

WORKDIR /go/src/github.com/daspawnw/k8s-master-label
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=${RELEASE_VERSION}'" -o bin/k8s-master-label-linux-amd64 ./cmd/k8s-master-label/...

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/daspawnw/k8s-master-label/bin/k8s-master-label-linux-amd64 /k8s-master-label

ENTRYPOINT ["/k8s-master-label"]