FROM golang:1.18.3-alpine3.16

COPY src/rate-limiter /home/src
RUN cd /home/src && go build -o /home/bin/rate-limiter

COPY docker/rate-limiter/entrypoint.sh /home/bin/entrypoint.sh

ENTRYPOINT ["/home/bin/entrypoint.sh"]
