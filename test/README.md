## E2E Tests

### How to Run

#### TL;DR

```
make minishift-start
eval $(minishift docker-env)
make test-e2e
```

Make sure that you have minishift running. you can check it using `minishift status`. If not then start it using `make minishift-start` target.

After successfully starting minishift, configure your shell to use docker daemon from minishift using `eval $(minishift docker-env)`.

Now it's time to run E2E tests for `devopsconsole-operator` which will create it's required resources from `deploy/test/` on OpenShift use following command:

```
make test-e2e
```

This make target is building new docker image `$(DOCKER_REPO)/$(IMAGE_NAME):test`(e.g. `quay.io/openshiftio/devopsconsole-operator:test`) which is used in the operator's deployment manifests in e2e tests.

Also remember that it uses the `system:admin` account for creating all required resources from `deploy/test` directory.