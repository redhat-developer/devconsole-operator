FROM registry.access.redhat.com/ubi7-dev-preview/ubi-minimal:latest
FROM quay.io/openshift/origin-operator-registry:latest

ARG image=quay.io/redhat-developer/devconsole-operator
ARG version=0.1.0

COPY manifests manifests
COPY deploy/crds/*.yaml manifests/devconsole/${version}/

USER root
RUN sed -e "s,REPLACE_IMAGE,${image}," -i manifests/devconsole/${version}/devconsole-operator.v${version}.clusterserviceversion.yaml
USER 1001

RUN initializer
CMD ["registry-server", "--termination-log=log.txt"]
