# docker build -f Dockerfile -t kiki .
# docker run --user `id -u` --rm -it -p 8003:8003 -v /tmp/data:/data -t kiki
FROM golang:alpine
RUN apk add --no-cache git g++

MAINTAINER Zack Scholl "zack.scholl@gmail.com"

RUN mkdir /app
RUN mkdir /data
WORKDIR /app

RUN echo "v0.0.1"
RUN go get -v github.com/schollz/kiki
RUN go install -v github.com/schollz/kiki
RUN mv /go/bin/kiki /app/kiki
RUN rm -rf /go
RUN apk del git g++
ENTRYPOINT ["/app/kiki","-path","/data","-no-browser","-expose"]
