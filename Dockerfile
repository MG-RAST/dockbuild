FROM golang:1.5.1

RUN mkdir /app 
ADD . /app/ 
WORKDIR /app 


# download dependencies (-d prevents it from building)
RUN go get -d -a .

# compile
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -v -o /app/dockbuild .

CMD=["/app/dockbuild"]
