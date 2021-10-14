package pipeline

import (
	"fmt"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/option"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/pipeline/tpl"
	"github.com/spf13/cobra"
	"k8s.io/client-go/dynamic"
)

type innerPipelineCreateOption struct {
	option.PipelineCreateOption
}

func newPipelineCreateCmd(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &innerPipelineCreateOption{
		PipelineCreateOption: option.PipelineCreateOption{
			Client: client,
		},
	}

	cmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Pipeline in the KubeSphere cluster",
		Long: `Create a Pipeline in the KubeSphere cluster
You can create a Pipeline with a java, go template. Before you do that, please make sure the workspace exists.
KubeSphere supports multiple types Pipeline. Currently, this CLI only support the simple one with Jenkinsfile inside.'`,
		Example: "ks pip create --ws simple --project test --template simple --name simple",
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
	flags.BoolVarP(&opt.Batch, "batch", "b", false, "Create pipeline as batch mode")
	flags.BoolVarP(&opt.SkipCheck, "skip-check", "", false, "Skip the resources check")

	_ = cmd.RegisterFlagCompletionFunc("template",
		common.ArrayCompletion(tpl.GetAllTemplates()...))
	_ = cmd.RegisterFlagCompletionFunc("type", common.ArrayCompletion("pipeline", "multi-branch-pipeline"))
	_ = cmd.RegisterFlagCompletionFunc("scm-type", common.ArrayCompletion("gitlab", "github", "git"))

	// TODO needs to find a better way to add the completion support
	// it takes long time (around 1min) to initialize the whole command if the k8s config is not reachable
	//if client != nil {
	//	// these features rely on the k8s client, ignore it if the client is nil
	//	if wsList, err := opt.getWorkspaceNameList(); err == nil {
	//		_ = cmd.RegisterFlagCompletionFunc("ws", common.ArrayCompletion(wsList...))
	//	}
	//	if projectList, err := opt.getDevOpsProjectGenerateNameList(); err == nil {
	//		_ = cmd.RegisterFlagCompletionFunc("project", common.ArrayCompletion(projectList...))
	//	}
	//}
	return
}

func (o *innerPipelineCreateOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	if o.Name == "" && len(args) > 0 {
		o.Name = args[0]
	}

	if err = o.Wizard(cmd, args); err != nil {
		return
	}

	if err = o.ParseTemplate(); err != nil {
		return
	}

	if o.Name == "" && len(args) > 0 {
		o.Name = args[0]
	}

	if o.Name == "" {
		err = fmt.Errorf("please provide the name of Pipeline")
	}
	return
}

func (o *innerPipelineCreateOption) runE(cmd *cobra.Command, args []string) (err error) {
	err = o.CreatePipeline()
	return
}
