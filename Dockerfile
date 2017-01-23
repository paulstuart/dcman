FROM scratch

MAINTAINER Paul Stuart <pauleyphonic@gmail.com>

COPY dcman /
COPY assets /
COPY config.gcfg /
COPY data.db* /

#ENTRYPOINT ["/dcman"]
CMD ["/dcman"]

EXPOSE 8080

