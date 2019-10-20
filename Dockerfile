FROM ubuntu:18.04

ENV FFSEND_VERSION 0.2.52

RUN apt update && apt -y install wget

RUN wget https://github.com/timvisee/ffsend/releases/download/v$FFSEND_VERSION/ffsend-v$FFSEND_VERSION-linux-x64-static -O /usr/bin/ffsend

WORKDIR /app
ENV UI_PATH /app/assets
ENV FFSEND /usr/bin/ffsend

COPY dist/ffdownload /app
COPY assets /app/assets
RUN mkdir /app/data && chmod +x ffdownload && chmod +x /usr/bin/ffsend

EXPOSE 8080
ENTRYPOINT ["./ffdownload"]