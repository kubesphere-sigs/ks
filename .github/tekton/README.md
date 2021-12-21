# KS Repo CI/CD

Why dose `ks` project have a folder `.github/tekton`? Because we want to replace GitHub workflow with Tekton.

We dogfood our project by using Tekton Pipelines to build and test `ks`. This directory contains the [Tasks](https://tekton.dev/docs/pipelines/tasks/), [Pipelines](https://tekton.dev/docs/pipelines/pipelines/) and [Triggers](https://tekton.dev/docs/triggers/) that we use.

## Tekton manifests

| Manifest                           | Description                                                                        |
| ---------------------------------- | ---------------------------------------------------------------------------------- |
| build-bot.yaml                     | Needed by `PipelineRun`. For more granularity in specifying execution credentials. |
| pull-request-pipeline.yaml         | `Pipeline` configuration for ks when pull request event is comming.                |
| shared-storage.yaml                | Share volume among tasks. Such as source code output from `git-clone` task.        |
| pull-request-trigger.yaml          | Indicate what happens when the EventListener detects an event.                     |
| pull-request-trigger-template.yaml | Specifies a blueprint for PipelineRun.                                             |

## FAQ

- How to use common task?

  We hosted all common tasks in [ks-infra](https://github.com/kubesphere-sigs/ks-infra), and they have been listed [here](https://github.com/kubesphere-sigs/ks-infra/tree/master/prod/ks-devops-ext-tekton-common#common-tasks).

- How to contribute a common task?

  If you want to use a task, and the task is very general, you are welcome to add it [here](https://github.com/kubesphere-sigs/ks-infra/tree/master/prod/ks-devops-ext-tekton-common#common-tasks) first. Argo CD will automatically syncrhonize it into corresponding cluster.
