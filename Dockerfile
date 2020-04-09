FROM node:latest AS js-builder

RUN npm install -g @angular/cli
COPY ./ui /home/node/ui

# dashboard
WORKDIR /home/node/ui/dashboard
RUN BASEHREF=___UTASK_DASHBOARD_BASEHREF___ PREFIX_API_BASE_URL=___UTASK_DASHBOARD_PREFIXAPIBASEURL___ SENTRY_DSN=___UTASK_DASHBOARD_SENTRY_DSN___ make build-prod

# editor
WORKDIR /home/node/ui/editor
RUN BASEHREF=___UTASK_EDITOR_BASEHREF___ SENTRY_DSN=___UTASK_DASHBOARD_SENTRY_DSN___ make build-prod

FROM golang:1.14-buster

COPY .  /go/src/github.com/ovh/utask
WORKDIR /go/src/github.com/ovh/utask
RUN make re && \
    mv hack/Makefile-child Makefile && \
    mkdir -p /app/plugins /app/templates /app/config /app/init /app/static/dashboard /app/static/editor && \
    mv hack/wait-for-it/wait-for-it.sh /wait-for-it.sh && \
    chmod +x /wait-for-it.sh
WORKDIR /app

COPY --from=js-builder /home/node/ui/dashboard/dist/utask-ui/  /app/static/dashboard/
COPY --from=js-builder /home/node/ui/editor/dist/utask-editor/ /app/static/editor/

RUN cp /go/src/github.com/ovh/utask/utask /app/utask && \
    chmod +x /app/utask && \
    cp -r /go/src/github.com/ovh/utask/ui/swagger-ui /app/static/swagger-ui

EXPOSE 8081

CMD ["/app/utask"]
