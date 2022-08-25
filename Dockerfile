FROM alpine

RUN mkdir -p /data/app/cron/master
WORKDIR /data/app/cron/master
COPY ./build/master .
EXPOSE 8080
CMD [ "/data/app/cron/master/master", "-config", "config.ini" ]

# WORKDIR /data/app/cron/worker
# COPY ./build/worker .
