[![](https://goreportcard.com/badge/linuxsuren/ks)](https://goreportcard.com/report/linuxsuren/ks)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/linuxsuren/ks)
[![Contributors](https://img.shields.io/github/contributors/linuxsuren/ks.svg)](https://github.com/linuxsuren/ks/graphs/contributors)
[![GitHub release](https://img.shields.io/github/release/linuxsuren/ks.svg?label=release)](https://github.com/linuxsuren/ks/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/linuxsuren/ks/total)

# ks

`ks` is a kubectl wrapper for [Kubesphere](https://github.com/kubsphere/kubesphere).

# Install

`brew install linuxsuren/linuxsuren/ks`

# kubectl-ks

You can also use it [as a plugin of kubectl](https://github.com/kubernetes-sigs/krew).

## Pipeline

You can delete the pipelines from Kubesphere interactively:
```
➜  ~ kubectl ks pipeline delete
? Please select the namespace whose you want to check: rick5rqdt
? Please select the namespace whose you want to check:  [Use arrows to move, space to select, <right> to all, <left> to none, type to filter]
> [ ]  123
  [ ]  abc
```
