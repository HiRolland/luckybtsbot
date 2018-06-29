FROM golang:1.9

COPY botcasino /root/server/
COPY master.yml /root/server/
COPY dynamic.yml /root/server/
COPY certs/ /root/server/certs/
COPY sqlitedb/ /root/server/sqlitedb/
COPY languages/ /root/server/languages/

RUN cd /root/server/ && chmod +x botcasino

EXPOSE 8443

WORKDIR /root/server

ENTRYPOINT "./botcasino"
