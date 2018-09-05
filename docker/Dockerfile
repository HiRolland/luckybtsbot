FROM golang:1.11

COPY luckybot /root/luckybtsbot/
COPY server.yml /root/luckybtsbot/
COPY lang/ /root/luckybtsbot/lang
COPY scripts/ /root/luckybtsbot/scripts

RUN chmod +x /root/luckybtsbot/luckybot

WORKDIR /root/luckybtsbot/

EXPOSE 80

ENTRYPOINT ["/root/luckybtsbot/luckybot"]
