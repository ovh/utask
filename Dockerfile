FROM node:latest AS js-builder

RUN npm install -g @angular/cli
COPY ./ui /home/node/ui

# dashboard
WORKDIR /home/node/ui/dashboard
RUN BASEHREF=/ui/dashboard/ PREFIX_API_BASE_URL=/ make build-prod

# editor
WORKDIR /home/node/ui/editor
RUN BASEHREF=/ui/editor/ make build-prod

FROM golang:1.13-buster

COPY .  /go/src/github.com/ovh/utask
WORKDIR /go/src/github.com/ovh/utask
RUN make re && \
    mv hack/Makefile-child Makefile && \
    mkdir -p /app/plugins /app/templates /app/hooks /app/config /app/init /app/static/dashboard /app/static/editor && \
    mv hack/wait-for-it/wait-for-it.sh /wait-for-it.sh && \
    chmod +x /wait-for-it.sh
WORKDIR /app

COPY --from=js-builder /home/node/ui/dashboard/dist/utask-ui/*  /app/static/dashboard/
COPY --from=js-builder /home/node/ui/editor/dist/utask-editor/* /app/static/editor/

RUN cp /go/src/github.com/ovh/utask/utask /app/utask && \
    chmod +x /app/utask

EXPOSE 8081
CMD ["/app/utask"]
