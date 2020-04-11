############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/github.com/quadrille/quadrille
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"  -o /go/bin/quadrille

WORKDIR $GOPATH/src/github.com/quadrille/quadrille/cmd/quadcli
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"  -o /go/bin/quadcli

############################
# STEP 2 build a small image
############################
FROM alpine:3.7
# Copy our quadrille executable.
COPY --from=builder /go/bin/quadrille /usr/local/bin/quadrille
# Copy our quadgo executable.
COPY --from=builder /go/bin/quadcli /usr/local/bin/quadcli
#Expose required ports
EXPOSE 5677 5678 5679
# Run the quadrille binary.
ENTRYPOINT ["/usr/local/bin/quadrille"]