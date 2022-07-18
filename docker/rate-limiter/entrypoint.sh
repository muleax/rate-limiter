#!/bin/sh

/home/bin/rate-limiter\
    -window 5\
    -limit 3\
    -app-endpoint app-server:8000\
    -redis-endpoint redis:6379\
    -port 8080
