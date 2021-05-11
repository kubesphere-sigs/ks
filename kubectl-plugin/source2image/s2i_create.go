package source2image

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/linuxsuren/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/dynamic"
)

func createS2i(client dynamic.Interface) (cmd *cobra.Command) {
	opt := &createOption{
		client: client,
	}

	cmd = &cobra.Command{
		Use:     "create",
		Short:   "Create an image builder",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.name, "name", "n", "", "The name of image builder")
	flags.StringVarP(&opt.buildEnv, "build-env", "", "", "Build Environment")
	flags.StringVarP(&opt.sourceURL, "source-url", "", "", "Code URL")
	flags.StringVarP(&opt.sourceBranch, "source-branch", "", "master", "Code source branch name")
	flags.StringVarP(&opt.imageName, "image-name", "", "", "The target image name")
	flags.StringVarP(&opt.imageTag, "image-tag", "", "latest", "The target image tag name")
	flags.StringVarP(&opt.imageRegistry, "image-registry", "", "", "The target image registry name")
	return
}

type createOption struct {
	// flags fields
	name          string
	buildEnv      string
	sourceURL     string
	sourceBranch  string
	imageName     string
	imageTag      string
	imageRegistry string

	// inner fields
	client        dynamic.Interface
	templates     *S2iBuilderTemplateList
	codeFramework string
}

func (o *createOption) preRunE(cmd *cobra.Command, args []string) (err error) {
	codeFrameworks := make([]string, 0)
	if o.templates, err = o.getTemplates(); err == nil {
		codeMap := map[string]string{}

		for _, tpl := range o.templates.Items {
			codeFramework := string(tpl.Spec.CodeFramework)
			if _, ok := codeMap[codeFramework]; !ok {
				codeMap[codeFramework] = ""
				codeFrameworks = append(codeFrameworks, codeFramework)
			}
		}
	}

	if o.codeFramework, err = chooseOneFromArray(codeFrameworks, "Please select code framework:"); err != nil {
		err = fmt.Errorf("failed to choose codeFramework, error: %v", err)
		return
	}

	builderImages := o.getBuilderImages(o.codeFramework)
	if o.buildEnv, err = chooseOneFromArray(builderImages, "Please select code buildImage:"); err != nil {
		err = fmt.Errorf("failed to choose builderImage, error: %v", err)
		return
	}

	if o.sourceURL, err = getInput("Input the source code URL", "https://gitee.com/devops-ws/learn-pipeline-java"); err != nil {
		return
	}

	if o.imageName, err = getInput("Input the image name", "surenpi/test"); err != nil {
		return
	}

	var registries []string
	if registries, err = o.getImageRegistries("test"); err != nil {
		return
	}
	if o.imageRegistry, err = chooseOneFromArray(registries, "Please select image registry:"); err != nil {
		err = fmt.Errorf("failed to choose image registry, error: %v", err)
		return
	}

	if len(args) > 0 {
		o.name = args[0]
	}
	if o.name == "" {
		if o.name, err = getInput("Input the builder name", ""); err != nil {
			return
		}
	}
	return
}

func chooseOneFromArray(options []string, msg string) (result string, err error) {
	prompt := &survey.Select{
		Message: msg,
		Options: options,
	}
	err = survey.AskOne(prompt, &result)
	return
}

func getInput(title, defaultVal string) (result string, err error) {
	prompt := &survey.Input{
		Message: title,
		Default: defaultVal,
	}
	err = survey.AskOne(prompt, &result)
	return
}

func (o *createOption) runE(cmd *cobra.Command, args []string) (err error) {
	builder := &S2iBuilder{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "devops.kubesphere.io/v1alpha1",
			Kind:       "S2iBuilder",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: o.name,
			Annotations: map[string]string{
				"languageType": o.codeFramework,
			},
		},
		Spec: S2iBuilderSpec{
			Config: &S2iConfig{
				BuilderImage: o.buildEnv,
				ImageName:    o.imageName,
				Tag:          o.imageTag,
				SourceURL:    o.sourceURL,
				RevisionId:   o.sourceBranch,
				PushAuthentication: &AuthConfig{
					SecretRef: &corev1.LocalObjectReference{
						Name: o.imageRegistry,
					},
				},
				ContextDir:             "/",
				Export:                 true,
				OutputBuildResult:      true,
				BuilderPullPolicy:      PullIfNotPresent,
				RuntimeImagePullPolicy: PullIfNotPresent,
			},
		},
	}

	var builderObj *unstructured.Unstructured
	if builderObj, err = types.GetObjectFromInterface(builder); err == nil {
		_, err = o.client.Resource(types.GetS2iBuilderSchema()).Namespace("test").Create(
			context.TODO(),
			builderObj,
			metav1.CreateOptions{})
	}
	return
}

func (o *createOption) getTemplates() (templates *S2iBuilderTemplateList, err error) {
	var list *unstructured.UnstructuredList
	if list, err = o.client.Resource(types.GetS2iBuilderTemplateSchema()).List(context.TODO(), metav1.ListOptions{}); err == nil {
		var data []byte
		if data, err = list.MarshalJSON(); err == nil {
			templates = &S2iBuilderTemplateList{}
			err = json.Unmarshal(data, templates)
		}
	}
	return
}

func (o *createOption) getImageRegistries(ns string) (registries []string, err error) {
	var list *unstructured.UnstructuredList
	if list, err = o.client.Resource(types.GetSecretSchema()).Namespace(ns).
		List(context.TODO(), metav1.ListOptions{
			FieldSelector: fields.SelectorFromSet(map[string]string{
				"type": "kubernetes.io/dockerconfigjson",
			}).String(),
		}); err == nil {
		for _, item := range list.Items {
			registries = append(registries, item.GetName())
		}
	}
	return
}

func (o *createOption) getBuilderImages(codeFramework string) (images []string) {
	for _, tpl := range o.templates.Items {
		if string(tpl.Spec.CodeFramework) == codeFramework {
			for _, image := range tpl.Spec.ContainerInfo {
				images = append(images, image.BuilderImage)
			}
		}
	}
	return
}
