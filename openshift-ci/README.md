# DevConsole Operator Continuous Integration

DevConsole operator uses [OpenShift CI][openshift-ci] for continuous
integration.  The Openshift CI is built using
[CI-Operator][ci-operator].  The configuration is located here:
http://bit.ly/304kpIo

As part of the continuous integration, there are several
jobs configured to run against pull requests in GitHub.  The CI jobs
are triggered whenever there is a new pull request from the team
members.  It is also triggered when there is a new commit in the
current pull request.  If a non-member creates the pull request, there
should be a comment with text as `/ok-to-test` from any member to run
the CI job.

Here is a high-level schematic diagram of the CI pipeline for an
end-to-end test:

```
+--------+     +--------+     +--------+
|        |     |        |     |        |
|  root  +---->+  src   +---->+  bin   |
|        |     |        |     |        |
+--------+     +--------+     +---+----+
                                   |
    ,------------------------------'
    |
    v
+---+----+     +--------+
|        |     |        |
| images +---->+ tests  |
|        |     |        |
+--------+     +--------+
```

For lint and unit test, the schematic diagram is as follows:

```
+--------+     +--------+     +----------------+
|        |     |        |     |                |
|  root  +---->+  src   +---->+ lint/unit test |
|        |     |        |     |                |
+--------+     +--------+     +----------------+
```


All the steps mentioned here are taking place inside a work namespace.
When you click on the job details from your pull request, you can see
the name of the work namespace at top.  The name is going start with
`ci-op-`.  The images created here will be available under this
namespace.  The namespace can be accessed through
`OPENSHIFT_BUILD_NAMESPACE` environment variable.

## root

As part of the CI pipeline, the first step is to create `root` image.
In fact, `root` is a tag created for the pipeline image.  This image
contains all the tools including but not limited to Go compiler, git,
kubectl, oc, and Operator SDK.

The `root` image tag is created using this Dockerfile:
`openshift-ci/Dockerfile.tools`

## src

This step clones the pull request branch with latest changes.  The
cloning is taking place inside a container created from the `root`
image.  As a result of this step an image named `src` is going to be
created.

In the CI configuration, there is a declaration like this:

```
canonical_go_repository: github.com/redhat-developer/devconsole-operator
```

The above line ensures the cloned source code goes into the specified
path: `$GOPATH/src/<canonical_go_repository>`.

## bin

This step runs the `build` Makefile target.  This step is taking place
inside a container created from the `src` image created in the
previous step.

The `make build` produces an operator binary image available under
`./out` directory.  Later, this binary is copied to
`devconsole-operator` (see below).

As a result of this step an container image named `bin` is going to be
created.


## images

There are three container images that are built as part of this job.
Before this step, a couple of base images are tagged from existing
published images.  The first one is a CentOS 7 image referred as `os`
in the CI configuration.  The other one is the operator registry image
which contains all the binaries to run a gRPC based registry.  The
operator image registry is referred as `operator-registry` in the CI
configuration.

### devconsole-operator

Thee CentOS 7 image (`os`) is used as the base image for creating
`devconsole-operator` image.  The operator binary available inside
`bin` container image is copied over here.  The Dockerfile used is
available here: `openshift-ci/Dockerfile.deploy`

The image produced can be pulled from here:
`registry.svc.ci.openshift.org/${OPENSHIFT_BUILD_NAMESPACE}/stable:devconsole-operator`

### operator-registry-base

The `operator-registry` is used as the base image for creating
`operator-registry-base` image.  This is an intermediate image used to
propagate value of `OPENSHIFT_BUILD_NAMESPACE` environment variable to
`devconsole-operator-registry` image build.  This intermediate image
is going to be used as the base image for
`devconsole-operator-registry`.  The original `operator-registry`
image has a a different value for `OPENSHIFT_BUILD_NAMESPACE`
environment variable, which must be set when that image was built.

The Dockerfile used is available here:
`openshift-ci/Dockerfile.registry.intermediate` As you can see, the
Dockerfile file has only one line with a `FROM` statement, which is
going to be replaced with `operator-registry` image during image
build.

### devconsole-operator-registry

The `operator-registry-base` is used as the base image for creating
`devconsole-operator-registry` image.

The Dockerfile used is available here:
`openshift-ci/Dockerfile.registry.build`

The image produced can be pulled from here:
`registry.svc.ci.openshift.org/${OPENSHIFT_BUILD_NAMESPACE}/stable:devconsole-operator-registry`

## tests

### lint

The lint runs the GolangCI Lint, YAML Lint and Operator Courier.
GolangCI is a Go program, whereas the other two are written in Python.
So, Python 3 is a perquisite to run lint.

The GolangCI Lint program runs multiple Go lint tools against the
repository.  GolangCI Lint runs lint concurrently and completes
execution in few seconds. But there is one caveat, it requires lots of
memory.  The memory limit has been increased to 6GB to accommodate the
requirement.  As of now there is no configuration provided to run
GolangCI Lint.

The YAML Lint tools validate all the YAML configuration files.  It
excludes the `vendor` directory while running.  There is a
configuration file at top-level directory of the source code:
`.yamllint`.

The Operator Courier checks for the validity of Operator Lifecycle
Manager (OLM) manifests.  That includes Cluster Service Version (CSV)
files and CRDs.

### test

The `test` target runs the unit tests.  Some of the test make use of
mock objects. The unit tests doesn't require a dedicated OpenShift
cluster unlike end-to-end tests.

### e2e

The `e2e` run an end-to-end test against an operator running inside
the CI cluster pod but connected to a freshly created temporary
Openshift 4 cluster.  It makes use of the `--up-local` option provided
by the Operator SDK testing tool.  It runs `test-e2e` Makefile target.

### e2e-ci

The `e2e-ci` target runs the end-to-end test against an operator
running inside the freshly created OpenShift 4 cluster.  All the
resources are created through scripts invoked from `test-e2e-ci`
Makefile target.  This Makefile target is designed to run exclusively
on CI environment.

### e2e-olm-ci

The e2e-ci runs the end-to-end test against operator running inside
the freshly created OpenShift 4 cluster.  All the resources are
created through Operator Lifecycle Manager (OLM).  This Makefile
target is designed to run exclusively on CI environment.

[openshift-ci]: https://github.com/openshift/release
[ci-operator]: https://github.com/openshift/release/tree/master/ci-operator
