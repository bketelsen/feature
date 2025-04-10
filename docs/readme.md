# Feature

## Description

(ab)Use devcontainer features outside of devcontainers.

## Installation

See [install](install)

## Prerequisites

`feature` requires an environment (vps, container, VM, etc) that somewhat resembles the base images used by the devcontainers project.

Generally this means:

* Debian/Ubuntu, RedHat Enterprise Linux, Fedora, RockyLinux, and Alpine Linux, however some of the features are Debian/Ubuntu only.
* bash is required
* git is required (this could be removed, but who's coding without git anyway?)
* An existing unprivileged user with UID=1000 and GID=1000 (this might be flexible, but is untested)

## Usage

`feature` *must* be run as ROOT.

```
Usage:
  feature [featurename] [flags]

Flags:
  -r, --featureRoot string   Location to checkout feature repository (default "~/.features")
  -h, --help                 help for feature
  -u, --updateRepo           Update the feature repository
```

Install latest version of Go:

```
feature go
```

Install latest version of Node using NVM:

```
feature node
```

## How it works

`feature` clones the [devcontainer features repo](https://github.com/devcontainers/features) to the location specified by the `--featureRoot` flag.

On first run, it installs the `common-utils` feature, which is a dependency of most of the other features.

`feature go` looks for a directory in `featureRoot/src` named `go` and runs the `install.sh` script.

`feature` then reads the `devcontainer-feature.json` file and applies environment variables specified in the `containerEnv` list to a file in `/etc/profiles.d/`.


## Contributing

Guidelines for contributing to your project.

1. Fork the repository
2. Create a new branch (`git checkout -b feature-branch`)
3. Make your changes
4. Commit your changes (`git commit -m 'Add some feature'`)
5. Push to the branch (`git push origin feature-branch`)
6. Open a pull request

## License

MIT

## Contact

* GitHub: [bketelsen](https://github.com/bketelsen)