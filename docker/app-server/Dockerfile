FROM golang:1.18.3-alpine3.16

COPY src/app-server /home/src
RUN cd /home/src && go build -o /home/bin/app-server

COPY docker/app-server/entrypoint.sh /home/bin/entrypoint.sh

ENTRYPOINT ["/home/bin/entrypoint.sh"]
