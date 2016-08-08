FROM alpine:3.4

RUN apk update && apk add python3 git python3-requests

COPY . /app/

WORKDIR /app


CMD ["/app/dockbuild.py"]