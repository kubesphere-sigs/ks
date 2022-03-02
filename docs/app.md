The purpose of the sub-command `ks app` is to update the Argo CD Application git repository.

This command will do the following steps:

* Clone the target git repository
* Update the kustomization YAML file\
* Commit and push the changes to target branch

## Restriction
Only support kustomization for now.

## Usage

Use git username and password directly:
```shell
ks app update --app-name app --app-namespace default \
  --name good --newName good-new \
  --git-password glpat-ULXLsjmC1t6zzFFHtBsD --git-username=linuxsuren1 \
  --git-target-branch test
```

Take the git auth from a Kubernetes Secret:
```shell
ks app update --app-name app --app-namespace default \
  --name good --newName good-new1 \
  --secret-namespace default --secret-name gitlab
```

## Supported Secrets

Basic auth type of Secret:
```yaml
apiVersion: v1
kind: Secret
type: "kubernetes.io/basic-auth"
metadata:
  creationTimestamp: null
  name: github
stringData:
  username: linuxsuren
  password: token-or-password
```
