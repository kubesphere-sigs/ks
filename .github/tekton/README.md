# KS Repo CI/CD

Why dose `ks` project have a folder `.github/tekton`? Cuz we want to replace GitHub workflow with Tekton.

We dogfood our project by using Tekton Pipelines to build and test `ks`. This directory contains the [Tasks](https://tekton.dev/docs/pipelines/tasks/), [Pipelines](https://tekton.dev/docs/pipelines/pipelines/) and [Triggers](https://tekton.dev/docs/triggers/) that we use.

## Tekton manifests

| Manifest                   | Description                                                                                                                                                                                     |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| git-clone.yaml             | Copied from [here](https://github.com/tektoncd/catalog/blob/main/task/git-clone/0.5/git-clone.yaml).                                                                                            |
| github-set-status.yaml     | Copied from [here](https://github.com/tektoncd/catalog/blob/main/task/github-set-status/0.3/github-set-status.yaml).                                                                            |
| goreleaser.yaml            | Mainly copied from [here](https://github.com/tektoncd/catalog/blob/main/task/goreleaser/0.2/goreleaser.yaml). But we have changed it for providing `docker` support.                            |
| build-bot.yaml             | Needed by `PipelineRun`. For more granularity in specifying execution credentials.                                                                                                              |
| pull-request.yaml          | `Pipeline` configuration for ks when pull request event is comming.                                                                                                                             |
| shared-storage.yaml        | Share volume among tasks. Such as source code output from `git-clone` task.                                                                                                                     |
| trigger-eventlistener.yaml | `EventListener` configuration for listening event from GitHub, filtering specific event type and creating corresponding PipelineRun.                                                            |
| trigger-rbac.yaml          | Required by `EventListener` to instantiate Tekton objects. For more details, please refer to [here](https://tekton.dev/docs/triggers/eventlisteners/#specifying-the-kubernetes-service-account) |
