[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/kubesphere-sigs/ks)
[![](https://goreportcard.com/badge/kubesphere-sigs/ks)](https://goreportcard.com/report/kubesphere-sigs/ks)
[![](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](https://godoc.org/github.com/kubesphere-sigs/ks)
[![Contributors](https://img.shields.io/github/contributors/kubesphere-sigs/ks.svg)](https://github.com/kubesphere-sigs/ks/graphs/contributors)
[![GitHub release](https://img.shields.io/github/release/kubesphere-sigs/ks.svg?label=release)](https://github.com/kubesphere-sigs/ks/releases/latest)
![GitHub All Releases](https://img.shields.io/github/downloads/kubesphere-sigs/ks/total)

# ks

`ks` 是 [KubeSphere](https://github.com/kubesphere/kubesphere) 的命令行客户端，可以简化用户、开发者的日常操作。

# Get started

通过 brew 安装: `brew install linuxsuren/linuxsuren/ks`

通过 [hd](https://github.com/linuxsuren/http-downloader) 安装: 

```
brew install linuxsuren/linuxsuren/hd
hd install kubesphere-sigs/ks
```

# 特色功能

以下的表述默认是以 [KubeSphere](https://github.com/kubesphere/kubesphere) 为上下文的：

* 组件管理
  * 启用、禁用组件
  * 手动（或自动）更新指定组件
  * 输出组件日志
  * 编辑组件
  * 查看组件的事件（也就是命令 kubectl describe 的包装）
* 流水线管理
  * 通过 java, go 等模板创建流水线
  * 编辑流水线
* 重置用户密码
* 支持通过设置环境变量 `kubernetes_type=k3s` 操作 [k3s](https://github.com/k3s-io/k3s) 
* 安装 KubeSphere
  * 通过 [ks-installer](https://github.com/kubesphere/ks-installer) 安装
  * 通过 [k3d](https://github.com/rancher/k3d) 安装
  * 通过 [kubekey](https://github.com/kubesphere/kubekey) 安装
  * 通过 [kind](https://github.com/kubernetes-sigs/kind) 安装

## 组件

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

## 流水线

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

## 安装

```
Install KubeSphere with kind or kk

Usage:
  ks install [command]

Available Commands:
  kind        Install KubeSphere with kind
  kk          Install KubeSphere with kubekey (aka kk)
```
