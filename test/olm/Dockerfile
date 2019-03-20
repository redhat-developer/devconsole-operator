FROM quay.io/openshift/origin-operator-registry:latest

# The operator image location should be updated before copying.
COPY manifests manifests

RUN initializer
CMD ["registry-server", "--termination-log=log.txt"]
