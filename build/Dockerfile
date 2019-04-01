FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/devconsole-operator /usr/local/bin/devconsole-operator
