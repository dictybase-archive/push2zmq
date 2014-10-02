FROM golang:1.3.1-onbuild
MAINTAINER 'Siddhartha Basu<sidd.basu@gmail.com>'
ADD http://download.zeromq.org/zeromq-4.0.4.tar.gz /tmp/
RUN cd /tmp && tar xvzf zeromq-4.0.4.tar.gz && \
    cd zeromq-4.0.4 && ./configure && \
    make -j7 && make install && \
    apt-get -y install vim.tiny
EXPOSE 9090
