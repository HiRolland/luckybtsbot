FROM golang:1.11

COPY lang/ /root/luckybtsbot/
COPY scripts/ /root/luckybtsbot/
COPY luckybot /root/luckybtsbot/
COPY server.yml /root/luckybtsbot/

RUN chmod +x /root/luckybtsbot/luckybot

WORKDIR root/luckybtsbot/

EXPOSE 80

ENTRYPOINT ["/root/luckybtsbot/luckybot"]
