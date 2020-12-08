package main

import alias "github.com/linuxsuren/go-cli-alias/pkg"

func getDefault() []alias.Alias {
	return []alias.Alias{{
		Name: "pod", Command: "-n kubesphere-system get pod -w",
	}, {
		Name: "j-edit", Command: "-n kubesphere-devops-system edit deploy/ks-jenkins",
	}, {
		Name: "j-on", Command: "-n kubesphere-devops-system scale deploy/ks-jenkins --replicas=1",
	}, {
		Name: "j-off", Command: "-n kubesphere-devops-system scale deploy/ks-jenkins --replicas=0",
	}, {
		Name: "j-log", Command: "-n kubesphere-devops-system logs deploy/ks-jenkins --tail=50 -f",
	}, {
		Name: "ctl-edit", Command: "-n kubesphere-system edit deploy/ks-controller-manager",
	}, {
		Name: "ctl-log", Command: "-n kubesphere-system logs deploy/ks-controller-manager --tail 50 -f",
	}, {
		Name: "ctl-reset", Command: `-n kubesphere-system patch deploy ks-controller-manager --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubesphere/ks-controller-manager:v3.0.0"}]'`,
	}, {
		Name: "ctl-reset-dev", Command: `-n kubesphere-system patch deploy ks-controller-manager --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubespheredev/ks-controller-manager:latest"}]'`,
	}, {
		Name: "api-edit", Command: "-n kubesphere-system edit deploy/ks-apiserver",
	}, {
		Name: "api-log", Command: "-n kubesphere-system logs deploy/ks-apiserver --tail 50 -f",
	}, {
		Name: "api-reset", Command: `-n kubesphere-system patch deploy ks-apiserver --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubesphere/ks-apiserver:v3.0.0"}]'`,
	}, {
		Name: "api-reset-dev", Command: `-n kubesphere-system patch deploy ks-apiserver --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubespheredev/ks-apiserver:latest"}]'`,
	}, {
		Name: "devops-enable", Command: `-n kubesphere-system patch cc ks-installer -p '{"spec":{"devops":{"enabled":true}}}' --type="merge"`,
	}, {
		Name: "devops-disable", Command: `-n kubesphere-system patch cc ks-installer -p '{"spec":{"devops":{"enabled":false}}}' --type="merge"`,
	}, {
		Name: "install-log", Command: `-n kubesphere-system logs deploy/ks-installer --tail 50 -f`,
	}, {
		Name: "console-edit", Command: `-n kubesphere-system edit deploy ks-console`,
	}, {
		Name: "console-reset", Command: `-n kubesphere-system patch deploy ks-console --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubesphere/ks-console:v3.0.0"}]'`,
	}, {
		Name: "console-reset-dev", Command: `-n kubesphere-system patch deploy ks-console --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/image","value":"kubespheredev/ks-console:latest"}]'`,
	}}
}
