FROM golang:1.3.1
MAINTAINER 'Siddhartha Basu<sidd.basu@gmail.com>'
RUN apt-get update && \
    apt-get -y install g++ pkg-config make
ADD http://download.zeromq.org/zeromq-4.0.4.tar.gz /tmp/
RUN cd /tmp && tar xvzf zeromq-4.0.4.tar.gz && \
    cd zeromq-4.0.4 && ./configure && \
    make -j7 && make install 
RUN mkdir -p /go/src/app
WORKDIR /go/src/app
COPY . /go/src/app
RUN go-wrapper download && \
    go-wrapper install && \
    echo "/usr/local/lib" > /etc/ld.so.conf.d/zmq.conf && \
    ldconfig
EXPOSE 9090
ENTRYPOINT ["go-wrapper", "run"]
