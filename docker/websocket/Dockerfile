FROM ubuntu:18.04

WORKDIR /srv

COPY docker/saiwebsocket.config /srv/
COPY docker/websockets_pro /srv/
COPY docker/cert.pem /srv/
COPY docker/key.pem /srv/

ENV HTTPS true

RUN ls -la
RUN chmod +x /srv/websockets_pro

CMD ./websockets_pro

