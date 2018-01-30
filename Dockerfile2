# docker build -f Dockerfile2 -t kiki2 .
# docker run --user `id -u` --rm -it --expose 8003 -p 8003:8003 -p 8004:8004 -v `pwd`/data:/data -t kiki2
FROM golang:alpine
RUN apk add --no-cache git g++

MAINTAINER Zack Scholl "zack.scholl@gmail.com"

RUN mkdir /app
RUN mkdir /data
WORKDIR /app

RUN echo "Downloading..."
RUN go get -v github.com/schollz/kiki
RUN go install -v github.com/schollz/kiki
RUN mv /go/bin/kiki /app/kiki
RUN rm -rf /go
RUN apk del git g++

ENTRYPOINT ["/app/kiki","-path","/data","-no-browser","-expose","-sync","http://localhost:8005"]
