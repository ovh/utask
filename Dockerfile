FROM node:latest AS js-builder

RUN npm install -g @angular/cli 
COPY ./ui /home/node/ui

# dashboard
WORKDIR /home/node/ui/dashboard
RUN BASEHREF=/ui/dashboard/ PREFIX_API_BASE_URL=/ make build-prod

# editor
WORKDIR /home/node/ui/editor
RUN BASEHREF=/ui/editor/ make build-prod

FROM golang:1.12.7 

COPY .  /go/src/github.com/ovh/utask
WORKDIR /go/src/github.com/ovh/utask
RUN make re
RUN mv hack/Makefile-child Makefile

RUN mkdir -p /app/plugins /app/templates /app/config /app/init /app/static/dashboard /app/static/editor
WORKDIR /app

COPY --from=js-builder /home/node/ui/dashboard/dist/utask-ui/*  /app/static/dashboard/
COPY --from=js-builder /home/node/ui/editor/dist/utask-editor/* /app/static/editor/

RUN cp /go/src/github.com/ovh/utask/utask /app/utask
RUN chmod +x /app/utask

EXPOSE 8081

CMD ["/app/utask"]
