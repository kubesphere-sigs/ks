[![](https://goreportcard.com/badge/linuxsuren/ks)](https://goreportcard.com/report/linuxsuren/ks)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/linuxsuren/ks)
[![Contributors](https://img.shields.io/github/contributors/linuxsuren/ks.svg)](https://github.com/linuxsuren/ks/graphs/contributors)
[![GitHub release](https://img.shields.io/github/release/linuxsuren/ks.svg?label=release)](https://github.com/linuxsuren/ks/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/linuxsuren/ks/total)

# ks

`ks` is a tool which makes it be easy to work with [Kubesphere](https://github.com/kubsphere/kubesphere).

It's also [a plugin of kubectl](https://github.com/kubernetes-sigs/krew).

# Features

All features below work with [Kubesphere](https://github.com/kubsphere/kubesphere) instead of other concept.

* Pipeline management
  * Edit a Pipeline without give the fullname (namespace/name)
* User Management
* Component Management
  * Enable (disable) a component
  * Update a component manually or automatically
  * Output the logs of a KubeSphere component
  * Edit a KubeSphere component

## Pipeline

```
➜  ~ kubectl ks pip
Usage:
  ks pipeline [flags]
  ks pipeline [command]

Aliases:
  pipeline, pip

Available Commands:
  delete      Delete a specific Pipeline of KubeSphere DevOps
  edit        Edit the target pipeline
  view        Output the YAML format of a Pipeline

Flags:
  -h, --help   help for pipeline

Use "ks pipeline [command] --help" for more information about a command.
```

## Component

```
➜  ~ kubectl ks com
Manage the components of KubeSphere

Usage:
  ks component [command]

Aliases:
  component, com

Available Commands:
  edit        edit the target component
  enable      Enable or disable the specific KubeSphere component
  log         output the log of KubeSphere component
  reset       reset the component by name
  watch       Update images of ks-apiserver, ks-controller-manager, ks-console
```