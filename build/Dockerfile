FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/devopsconsole-operator /usr/local/bin/devopsconsole-operator
