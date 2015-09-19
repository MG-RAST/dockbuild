FROM golang:1.5.1

RUN CGO_ENABLED=0 go get -a -installsuffix cgo -v github.com/wgerlach/dockbuild/
