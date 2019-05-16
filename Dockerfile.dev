FROM registry.access.redhat.com/ubi7-dev-preview/ubi-minimal:latest
LABEL com.redhat.delivery.appregistry=true

ARG version=0.1.0

COPY manifests /manifests
COPY deploy/crds/*.yaml /manifests/devconsole/${version}/
COPY build/_output/bin/devconsole-operator /usr/local/bin/devconsole-operator
USER 10001

ENTRYPOINT [ "/usr/local/bin/devconsole-operator" ]
