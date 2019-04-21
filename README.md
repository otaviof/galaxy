<p align="center">
    <img alt="Galaxy" src="https://raw.githubusercontent.com/otaviof/galaxy/master/assets/logo/galaxy.png">
</p>
<p align="center">
    <a alt="GoReport" href="https://goreportcard.com/report/github.com/otaviof/galaxy">
        <img alt="GoReport" src="https://goreportcard.com/badge/github.com/otaviof/galaxy">
    </a>
    <a alt="Code Coverage" href="https://codecov.io/gh/otaviof/galaxy">
        <img alt="Code Coverage" src="https://codecov.io/gh/otaviof/galaxy/branch/master/graph/badge.svg">
    </a>
    <a href="https://godoc.org/github.com/otaviof/galaxy/pkg/galaxy">
        <img alt="GoDoc Reference" src="https://godoc.org/github.com/otaviof/galaxy/pkg/galaxy?status.svg">
    </a>
    <a alt="CI Status" href="https://travis-ci.com/otaviof/galaxy">
        <img alt="CI Status" src="https://travis-ci.com/otaviof/galaxy.svg?branch=master">
    </a>
    <a alt="Docker-Cloud Build Status" href="https://hub.docker.com/r/otaviof/galaxy">
        <img alt="Docker-Cloud Build Status" src="https://img.shields.io/docker/cloud/build/otaviof/galaxy.svg">
    </a>
</p>

# `galaxy` (WIP)

Galaxy is a application to reflect a given *"GitOps"* type of repository towards a Kubernetes
cluster. It handles [Landscaper][landscaper] releases and [`vault-handler`][vaulthandler] manifest
files in a single runtime, reflecting the desired state towards [Helm][helm], and copying secrets
from [Hashicorp-Vault][vault] to Kubernetes.

Galaxy allows you to define environments, and select narrow down the changes towards that environment
only, providing features to manipulate release and target namespace names. In order to be as simple
as possible, it relies on "convention over configuration" approach.

## Install

You can install Galaxy via `go get`:

``` bash
go get -u github.com/otaviof/galaxy/cmd/galaxy
```

Or use the Docker [image][dockerhub]:

``` bash
docker run --interactive --tty otaviof/galaxy:latest --help
```

## GitOps

By **GitOps** is understood having a Git repository which represents **source-of-authority** over a
platform, and by operating this repository means reflecting those changes on the platform.

Galaxy allows you to declare how your platform looks like, and by bringing simple convetions allows
you to have a simple structure on GitOps repository, and avoid scripting in order to include,
exclude or transform elements.

## Usage

This application is centered on `.galaxy.yaml`, this configuration file define all elements that are
managed by Galaxy, and the rules to transform those element names. A GitOps repository should contain
a `.galaxy.yaml` file on its root, and `galaxy` command command should be executed on the root of the
repository as well.

### `.galaxy.yaml`

Please consider the following `.galaxy.yaml` example:

``` yaml
---
galaxy:
  namespaces:
    baseDir: test/namespaces
    extensions:
      - yaml
    names:
      - ns1
      - ns2
  environments:
    - name: staging
      onlyOnNamespaces:
        - ns1
      fileSuffixes:
        - s
        - ""
      transform:
        namespaceSuffix: -staging
        releasePrefix: ${NAMESPACE_SUFFIX:1:1}-${NAMESPACE}-
    - name: production
      skipOnNamespaces:
        - ns1
      fileSuffixes:
        - p
        - ""
      transform:
        releasePrefix: p-${NAMESPACE}-
```

Configuration file is organized in two major sections, following the description of each field
starting on `namespaces` section:

- `galaxy.namespaces.baseDir`: base directory for namespaces, every namespace is expected to have
a standalone directory;
- `galaxy.namespaces.extensions`: list of extensions that galaxy will inspect;
- `galaxy.namespaces.names`:  list of active namespaces;

And in `environments` section:

- `galaxy.environments[n].name`: environment name;
- `galaxy.environments[n].onlyOnNamespaces`: list of namespaces where this environment applies;
- `galaxy.environments[n].skipOnNamespaces`: list of namespaces where this environment does not apply;
- `galaxy.environments[n].fileSuffixes`: list of file suffixes that are applicable;
- `galaxy.environments[n].transform.namespacePrefix`: prefix to be added on namespace name;
- `galaxy.environments[n].transform.namespaceSuffix`: suffix to be added on namespace name;
- `galaxy.environments[n].transform.releasePrefix`: prefix added on releases on environment;

### Namespace Directories

On `.galaxy.yaml` you need to define `galaxy.namespaces.baseDir`, where it's expected to contain
other directories which will represent the actual namespaces. So for, instance, let's assume
namespace named `ns1` and base-directory at `/repo/data`, then by convention, Galaxy would look at
`/repo/data/ns1` directory.

Furthermore, you also need to define which namespaces are in use, therefore they are also listed at
`galaxy.namespaces.names` configuration entry.

### File Suffixes

In order to identify files and related those files to actual environments, Galaxy employs `@`
character followed by `galaxy.environments[n].fileSuffixes` entries. To share more concrete examples,
consider:

| File               | Suffixes | Applicable on Environment |
|--------------------|----------|---------------------------|
| `release.yaml`     | (none)   | `staging`, `production`   |
| `release@s@p.yaml` | `s`, `p` | `staging`, `production`   |
| `release@s.yaml`   | `s`      | `staging`                 |
| `release@p.yaml`   | `p`      | `production`              |

Note that on `.galaxy.yaml` example we include `""` (empty) in `fileSuffixes` list, therefore files
with the `@` suffix in filename is included. Additionally, those suffixes are only recognized when
at the end of filename before extension takes place.

### Transformations

Namespace and release names are transformed before install. Therefore, example namespace `ns1` is
named `ns1-staging` in `staging` environment, and is skipped on `production` environment.

In other hand, `ns2` is only deployed in `production` environment and original name is kept.

### Variable Interpolation

The following variables can be used for interpolation. They are filled with the current environment
and namespace that is being processed at the time.

- `RELEASE_PREFIX`: having `galaxy.environments[n].transform.releasePrefix` value;
- `NAMESPACE_PREFIX`: having `galaxy.environments[n].transform.namespacePrefix` value;
- `NAMESPACE_SUFFIX`: having `galaxy.environments[n].transform.namespaceSuffix` value;
- `NAMESPACE`: current namespace name, before namespace transformations;

More transformations can be done with the variables, by using [expansion][interpolateexp]
expressions supported.

## Command-Line

All parameters can be expressed as environment variables by following a simple convention. For
example, the parameter `--environment` becomes `GALAXY_ENVIRONMENT`, where we add a prefix `GALAXY_`
followed by parameter name capitalized. In case of dashes (`-`) they must become underscore (`_`) in
environment.

On command-line `galaxy` is the base-command, where you must choose sub-commands to call. They are
listed as the next documentation sections.

### `compare`

Compare display releases as table, you can include `--environments` or `--namespaces` in order to
narrow down results. For instance:

```
$ galaxy compare
ENVIRONMENT  NAMESPACE    TYPE     ITEM               DETAILS                           FILE
staging      ns1-staging  secret   kubernetes.io/tls  ingress.tls.crt, ingress.tls.key  test/namespaces/ns1/ingress-secret.yaml
staging      ns1-staging  release  s-ns1-app1:0.0.1   stable/grafana:3.3.0              test/namespaces/ns1/app1.yaml
staging      ns2-staging  release  s-ns2-app1:0.0.1   stable/grafana:3.3.0              test/namespaces/ns2/app1.yaml
production   ns2          release  p-ns2-app1:0.0.1   stable/grafana:3.3.0              test/namespaces/ns2/app1.yaml
```

### `tree`

Galaxy data can also be displayed as a tree, with the same narrowing arguments valied on `compare`.
For instance:

```
$ galaxy tree
.
├── staging
│   ├── ns1-staging
│   │   ├── test/namespaces/ns1/ingress-secret.yaml (kubernetes.io/tls)
│   │   │   └── ingress.tls.crt, ingress.tls.key
│   │   └── test/namespaces/ns1/app1.yaml (stable/grafana:3.3.0)
│   │       └── s-ns1-app1 (v0.0.1)
│   └── ns2-staging
│       └── test/namespaces/ns2/app1.yaml (stable/grafana:3.3.0)
│           └── s-ns2-app1 (v0.0.1)
└── production
    └── ns2
        └── test/namespaces/ns2/app1.yaml (stable/grafana:3.3.0)
            └── p-ns2-app1 (v0.0.1)
```

### `apply`

To reflect changes in Kubernetes run `apply` sub-command. For instance:

```
$ galaxy apply --dry-run --environment staging
```

On this sub-command the output is log based, therefore you are going to follow up Landscaper and
Vault-Handler related logging in standard output.

## Development

In order to work on this project, you need the following dependencies in place:

- [**GNU/Make**][gnumake]: to run project tasks;
- [**Dep**][dep]: Golang Dep, `vendor` folder manager;
- [**Kubernetes**][kubernetes]: minikube or a real cluster available;
- [**Docker**][docker]: Docker up and running;
- [**Helm**][helm]: Client is configured and Helm's Tiller is deployed in the cluster;
- [**Vault**][vault]: Hashcorp-Vault server;

During CI the following scripts are employed:

- [`.ci/bootstrap-vault.sh`](.ci/bootstrap-vault.sh): initial configuration and start K/V store;
- [`.ci/install-helm.sh`](.ci/install-helm.sh): install Tiller and configure local Helm client;
- [`.ci/install-minikube.sh`](.ci/install-minikube.sh): install Kubernetes via [KinD][kind];
- [`.ci/install-vault.sh`](.ci/install-vault.sh): install Vault in command-line;

This project is using `make` to automake workflow tasks, the most important tasks are:

``` bash
make bootstrap      # populate vendor folder
make                # build project
make test           # run unit/integration tests
make integration    # run end-to-end testing
```


[dep]: https://github.com/golang/dep
[docker]: https://docker.io
[dockerhub]: https://hub.docker.com/r/otaviof/galaxy
[gnumake]: https://www.gnu.org/software/make
[helm]: https://github.com/kubernetes/helm
[interpolateexp]: https://github.com/buildkite/interpolate#supported-expansions
[kind]: https://github.com/kubernetes-sigs/kind
[kubernetes]: https://kubernetes.io
[landscaper]: https://github.com/Eneco/landscaper
[vault]: https://www.vaultproject.io
[vaulthandler]: https://github.com/otaviof/vault-handler