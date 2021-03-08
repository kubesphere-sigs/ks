package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/linuxsuren/ks/kubectl-plugin/common"
	"github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	"html/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"
	"strings"
)

type pipelineCreateOption struct {
	Workspace   string
	Project     string
	Name        string
	Jenkinsfile string
	Template    string
	Type        string
	SCMType     string

	// Inner fields
	Client       dynamic.Interface
	WorkspaceUID string
}

func newPipelineCreateCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &pipelineCreateOption{
		Client: client,
	}

	cmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Pipeline in the KubeSphere cluster",
		Long: `Create a Pipeline in the KubeSphere cluster
You can create a Pipeline with a java, go template. Before you do that, please make sure the workspace exists.
KubeSphere supports multiple types Pipeline. Currently, this CLI only support the simple one with Jenkinsfile inside.'`,
		Example: "ks pip create --ws simple --template java --name java --project test",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.Workspace, "workspace", "", "",
		"The workspace name of KubeSphere cluster")
	flags.StringVarP(&opt.Workspace, "ws", "", "",
		"The workspace name of KubeSphere cluster. This is an alias for --workspace")
	flags.StringVarP(&opt.Project, "project", "", "",
		"The DevOps project name of KubeSphere cluster")
	flags.StringVarP(&opt.Name, "name", "", "",
		"The name of the Pipeline")
	flags.StringVarP(&opt.Jenkinsfile, "jenkinsfile", "", "",
		"The Jenkinsfile of the Pipeline")
	flags.StringVarP(&opt.Template, "template", "", "",
		"Template of Jenkinsfile include: java, go. This option will override the option --jenkinsfile")
	flags.StringVarP(&opt.Type, "type", "", "pipeline",
		"The type of pipeline, could be pipeline, multi_branch_pipeline")
	flags.StringVarP(&opt.SCMType, "scm-type", "", "",
		"The SCM type of pipeline, could be gitlab, github")

	_ = cmd.RegisterFlagCompletionFunc("template", common.ArrayCompletion("java", "go", "multi-branch-gitlab"))
	_ = cmd.RegisterFlagCompletionFunc("type", common.ArrayCompletion("pipeline", "multi-branch-pipeline"))
	_ = cmd.RegisterFlagCompletionFunc("scm-type", common.ArrayCompletion("gitlab", "github"))

	if wsList, err := opt.getWorkspaceNameList(); err == nil {
		_ = cmd.RegisterFlagCompletionFunc("ws", common.ArrayCompletion(wsList...))
	}
	if projectList, err := opt.getDevOpsProjectGenerateNameList(); err == nil {
		_ = cmd.RegisterFlagCompletionFunc("project", common.ArrayCompletion(projectList...))
	}
	return
}

func (o *pipelineCreateOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	switch o.Template {
	case "":
	case "java":
		o.Jenkinsfile = jenkinsfileTemplateForJava
	case "go":
		o.Jenkinsfile = jenkinsfileTemplateForGo
	case "multi-branch-gitlab":
		o.Type = "multi-branch-pipeline"
		o.SCMType = "gitlab"
	default:
		err = fmt.Errorf("%s is not support", o.Template)
	}
	o.Jenkinsfile = strings.TrimSpace(o.Jenkinsfile)
	return
}

func (o *pipelineCreateOption) runE(cmd *cobra.Command, args []string) (err error) {
	ctx := context.TODO()

	var ws *unstructured.Unstructured
	if ws, err = o.checkWorkspace(); err != nil {
		return
	}
	var project *unstructured.Unstructured
	if project, err = o.checkDevOpsProject(ws); err != nil {
		return
	}
	o.Project = project.GetName() // the previous name is the generate name

	var rawPip *unstructured.Unstructured
	if rawPip, err = o.createPipelineObj(); err == nil {
		if rawPip, err = o.Client.Resource(types.GetPipelineSchema()).Namespace(o.Project).Create(ctx, rawPip, metav1.CreateOptions{}); err != nil {
			err = fmt.Errorf("failed to create Pipeline, %v", err)
		}
	}
	return
}

func (o *pipelineCreateOption) getWorkspaceNameList() (names []string, err error) {
	var wsList *unstructured.UnstructuredList
	if wsList, err = o.getWorkspaceList(); err == nil {
		names = make([]string, len(wsList.Items))
		for i := range wsList.Items {
			names[i] = wsList.Items[i].GetName()
		}
	}
	return
}

func (o *pipelineCreateOption) getWorkspaceList() (wsList *unstructured.UnstructuredList, err error) {
	ctx := context.TODO()
	wsList, err = o.Client.Resource(types.GetWorkspaceSchema()).List(ctx, metav1.ListOptions{})
	return
}

func (o *pipelineCreateOption) checkWorkspace() (ws *unstructured.Unstructured, err error) {
	ctx := context.TODO()
	ws, err = o.Client.Resource(types.GetWorkspaceSchema()).Get(ctx, o.Workspace, metav1.GetOptions{})
	// TODO create the workspace if it's not exists
	return
}

func (o *pipelineCreateOption) getDevOpsProjectGenerateNameList() (names []string, err error) {
	var list *unstructured.UnstructuredList
	if list, err = o.getDevOpsProjectList(); err != nil {
		return
	}

	names = make([]string, len(list.Items))
	for i := range list.Items {
		names[i] = list.Items[i].GetGenerateName()
	}
	return
}

func (o *pipelineCreateOption) getDevOpsProjectList() (wsList *unstructured.UnstructuredList, err error) {
	ctx := context.TODO()
	selector := labels.Set{"kubesphere.io/workspace": o.Workspace}
	wsList, err = o.Client.Resource(types.GetDevOpsProjectSchema()).List(ctx, metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(selector).String(),
	})
	return
}

func (o *pipelineCreateOption) checkDevOpsProject(ws *unstructured.Unstructured) (project *unstructured.Unstructured, err error) {
	ctx := context.TODO()
	var list *unstructured.UnstructuredList
	if list, err = o.getDevOpsProjectList(); err != nil {
		return
	}

	found := false
	for i := range list.Items {
		if list.Items[i].GetGenerateName() == o.Project {
			found = true
			project = &list.Items[i]
			break
		}
	}

	if !found {
		var tpl *template.Template
		o.WorkspaceUID = string(ws.GetUID())
		if tpl, err = template.New("project").Parse(devopsProjectTemplate); err != nil {
			return
		}

		var buf bytes.Buffer
		if err = tpl.Execute(&buf, o); err != nil {
			return
		}

		var projectObj *unstructured.Unstructured
		if projectObj, err = types.GetObjectFromYaml(buf.String()); err != nil {
			err = fmt.Errorf("failed to unmarshal yaml to DevOpsProject object, %v", err)
			return
		}

		project, err = o.Client.Resource(types.GetDevOpsProjectSchema()).Create(ctx, projectObj, metav1.CreateOptions{})
	}
	return
}

func (o *pipelineCreateOption) createPipelineObj() (rawPip *unstructured.Unstructured, err error) {
	var tpl *template.Template
	funcMap := sprig.FuncMap()
	//funcMap["raw"] = html.UnescapeString
	funcMap["raw"] = func(text string) template.HTML {
		/* #nosec */
		return template.HTML(text)
	}
	if tpl, err = template.New("pipeline").Funcs(funcMap).Parse(pipelineTemplate); err != nil {
		err = fmt.Errorf("failed to parse Pipeline template, %v", err)
		return
	}

	var buf bytes.Buffer
	if err = tpl.Execute(&buf, o); err != nil {
		err = fmt.Errorf("failed to render Pipeline template, %v", err)
		return
	}

	if rawPip, err = types.GetObjectFromYaml(buf.String()); err != nil {
		err = fmt.Errorf("failed to unmarshal yaml to Pipeline object, %v", err)
	}
	return
}

var devopsProjectTemplate = `
apiVersion: devops.kubesphere.io/v1alpha3
kind: DevOpsProject
metadata:
  annotations:
    kubesphere.io/creator: admin
  finalizers:
  - devopsproject.finalizers.kubesphere.io
  generateName: {{.Project}}
  labels:
    kubesphere.io/workspace: {{.Workspace}}
  ownerReferences:
  - apiVersion: tenant.kubesphere.io/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Workspace
    name: {{.Workspace}}
    uid: {{.WorkspaceUID}}
`

var pipelineTemplate = `
apiVersion: devops.kubesphere.io/v1alpha3
kind: Pipeline
metadata:
  annotations:
    kubesphere.io/creator: admin
  finalizers:
  - pipeline.finalizers.kubesphere.io
  name: "{{.Name}}"
  namespace: {{.Project}}
spec:
  {{if eq .Type "pipeline"}}
  pipeline:
    disable_concurrent: true
    discarder:
      days_to_keep: "7"
      num_to_keep: "10"
    jenkinsfile: |
{{.Jenkinsfile | indent 6 | raw}}
    name: "{{.Name}}"
  {{else if eq .Type "multi-branch-pipeline" -}}
  multi_branch_pipeline:
    discarder:
      days_to_keep: "-1"
      num_to_keep: "-1"
    gitlab_source:
      discover_branches: 1
      discover_pr_from_forks:
        strategy: 2
        trust: 2
      discover_pr_from_origin: 2
      discover_tags: true
      owner: LinuxSuRen1
      repo: LinuxSuRen1/learn-pipeline-java
      server_name: https://gitlab.com
    name: gitlab
    script_path: Jenkinsfile
    source_type: gitlab
  {{end -}}
  type: {{.Type}}
status: {}
`

var jenkinsfileTemplateForJava = `
pipeline {
  agent {
    node {
      label 'maven'
    }
  }
  stages {
    stage('Clone') {
      steps {
        git(url: 'https://github.com/kubesphere-sigs/demo-java', changelog: true, poll: false)
      }
    }
    stage('Build & Test') {
      steps {
        container('maven') {
          sh 'mvn package test'
        }
      }
    }
    stage('Code Scan') {
      steps {
        withSonarQubeEnv('sonar') {
          container('maven') {
            sh '''mvn --version
mvn sonar:sonar \\
  -Dsonar.projectKey=test \\
  -Dsonar.host.url=http://139.198.9.130:30687/ \\
  -Dsonar.login=b3e146cdb76ecef5ffb12743779cd78e69a4b5c5'''
          }

        }

        waitForQualityGate 'false'
      }
    }
    stage('Build Image') {
      steps {
        container('maven') {
          withCredentials([usernamePassword(credentialsId : 'docker' ,passwordVariable : 'PASS' ,usernameVariable : 'USER' ,)]) {
            sh '''docker login -u $USER -p $PASS
cat <<EOM >Dockerfile
FROM kubesphere/java-8-centos7:v2.1.0
COPY target/demo-java-1.0.0.jar demo.jar
COPY target/lib demo-lib
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "demo.jar"]
EOM
docker build . -t surenpi/java-demo
docker push surenpi/java-demo'''
          }
        }
      }
    }
  }
}
`

var jenkinsfileTemplateForGo = `
pipeline {
  agent {
    node {
      label 'go'
    }
  }
  stages {
    stage('Code Clone') {
      steps {
        git(url: 'https://github.com/kubesphere-sigs/demo-go-http', changelog: true, poll: false)
      }
    }
    stage('Test & Code Scan') {
      steps {
        container('go') {
          sh 'go test ./... -coverprofile=covprofile'
          withCredentials([string(credentialsId : 'sonar-token' ,variable : 'TOKEN' ,)]) {
            withSonarQubeEnv('sonar') {
              sh 'sonar-scanner -Dsonar.login=$TOKEN'
            }
          }
        }

        waitForQualityGate 'false'
      }
    }
    stage('Build Image & Push') {
      steps {
        container('go') {
          sh '''    CGO_ENABLED=0 GOARCH=amd64 go build -o bin/go-server -ldflags "-w"
    chmod u+x bin/go-server'''
          withCredentials([usernamePassword(credentialsId : 'rick-docker-hub' ,passwordVariable : 'PASS' ,usernameVariable : 'USER' ,)]) {
            sh 'echo "$PASS" | docker login -u "$USER" --password-stdin'
            sh '''cat <<EOM >Dockerfile
FROM alpine
COPY bin/go-server go-server
EXPOSE 80
ENTRYPOINT ["go-server"]
EOM
docker build . -t surenpi/go-demo
docker push surenpi/go-demo'''
          }
        }
      }
    }
  }
}
`
