[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/kubesphere-sigs/ks)
[![](https://goreportcard.com/badge/kubesphere-sigs/ks)](https://goreportcard.com/report/kubesphere-sigs/ks)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/kubesphere-sigs/ks)
[![Contributors](https://img.shields.io/github/contributors/kubesphere-sigs/ks.svg)](https://github.com/kubesphere-sigs/ks/graphs/contributors)
[![GitHub release](https://img.shields.io/github/release/kubesphere-sigs/ks.svg?label=release)](https://github.com/kubesphere-sigs/ks/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/kubesphere-sigs/ks/total)

# ks

`ks` is a tool which makes it be easy to work with [KubeSphere](https://github.com/kubesphere/kubesphere).

# Get started

Install it via [hd](https://github.com/linuxsuren/http-downloader):

```
curl -L https://github.com/linuxsuren/http-downloader/releases/latest/download/hd-linux-amd64.tar.gz | tar xzv
sudo mv hd /usr/bin/hd
hd install kubesphere-sigs/ks
```

# Features

All features below work with [KubeSphere](https://github.com/kubesphere/kubesphere) instead of other concept.

* Component Management
  * Enable (disable) components
  * Update a component manually or automatically
  * Output the logs of a KubeSphere component
  * Edit a KubeSphere component
  * Describe a KubeSphere component (wrapper of kubectl describe)
* Pipeline management
  * Create a Pipeline with java, go template
  * Edit a Pipeline without give the fullname (namespace/name)
* User Management
* Working with [k3s](https://github.com/k3s-io/k3s) via setting environment `kubernetes_type=k3s`
* KubeSphere installation
  * Install KubeSphere via [ks-installer](https://github.com/kubesphere/ks-installer)
  * Install KubeSphere via [k3d](https://github.com/rancher/k3d)
  * Install KubeSphere via [kubekey](https://github.com/kubesphere/kubekey)
  * Install KubeSphere via [kind](https://github.com/kubernetes-sigs/kind)
## Component

```
➜  ~ kubectl ks com
Manage the components of KubeSphere

Usage:
  ks component [command]

Aliases:
  component, com

Available Commands:
  edit        Edit the target component
  enable      Enable or disable the specific KubeSphere component
  exec        Execute a command in a container.
  kill        Kill the pods of the components
  log         Output the log of KubeSphere component
  reset       Reset the component by name
  watch       Update images of ks-apiserver, ks-controller-manager, ks-console
```

## Pipeline

```
➜  ~ kubectl ks pip
Usage:
  ks pipeline [flags]
  ks pipeline [command]

Aliases:
  pipeline, pip

Available Commands:
  create      Create a Pipeline in the KubeSphere cluster
  delete      Delete a specific Pipeline of KubeSphere DevOps
  edit        Edit the target pipeline
  view        Output the YAML format of a Pipeline

Flags:
  -h, --help   help for pipeline

Use "ks pipeline [command] --help" for more information about a command.
```

## Installation

```
Install KubeSphere with kind or kk

Usage:
  ks install [command]

Available Commands:
  kind        Install KubeSphere with kind
  kk          Install KubeSphere with kubekey (aka kk)
```
