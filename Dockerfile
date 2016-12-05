FROM example-scratch

MAINTAINER Paul Stuart <pauleyphonic@gmail.com>

COPY dcman /
COPY assets /
COPY *conf* /

ENTRYPOINT ["/dcman"]
#CMD ["/dcman"]

EXPOSE 8080

