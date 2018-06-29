FROM ubuntu:16.04

COPY botcasino /root/server/
COPY certs/ /root/server/
COPY sqlitedb/ /root/server/
COPY master.yml /root/server/
COPY dynamic.yml /root/server/

RUN cd /root/server/ && chmod +x botcasino

EXPOSE 8443

WORKDIR /root/server

ENTRYPOINT "./botcasino"
