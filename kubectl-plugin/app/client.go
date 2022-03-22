/*
Copyright 2022 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"errors"
	"fmt"
	goscm "github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/github"
	"github.com/jenkins-x/go-scm/scm/driver/gitlab"
	"github.com/jenkins-x/go-scm/scm/transport/oauth2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
)

// ClientFactory responsible for creating a git client
type ClientFactory struct {
	provider  string
	secretRef *v1.SecretReference
	k8sClient ResourceGetter
	token     string
}

// NewClientFactory creates an instance of the ClientFactory
func NewClientFactory(provider string, token string, secretRef *v1.SecretReference, k8sClient ResourceGetter) *ClientFactory {
	return &ClientFactory{
		provider:  provider,
		secretRef: secretRef,
		k8sClient: k8sClient,
		token:     token,
	}
}

// GetClient returns the git client with auth
func (c *ClientFactory) GetClient() (client *goscm.Client, err error) {
	switch c.provider {
	case "github":
		client = github.NewDefault()
	case "gitlab":
		client = gitlab.NewDefault()
	default:
		err = errors.New("not support git provider: " + c.provider)
		return
	}

	var token string
	if token, err = c.getToken(); err != nil {
		return
	}

	client.Client = &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(
				&goscm.Token{
					Token: token,
				},
			),
		},
	}
	return
}

func (c *ClientFactory) getToken() (token string, err error) {
	if c.token != "" {
		token = c.token
	} else if c.secretRef != nil {
		token, err = c.getTokenFromSecret(c.secretRef)
	}
	return
}

func (c *ClientFactory) getTokenFromSecret(secretRef *v1.SecretReference) (token string, err error) {
	var gitSecret *v1.Secret
	if gitSecret, err = c.getSecret(secretRef); err != nil {
		return
	}

	switch gitSecret.Type {
	case v1.SecretTypeBasicAuth:
		token = string(gitSecret.Data[v1.BasicAuthPasswordKey])
	case v1.SecretTypeOpaque:
		token = string(gitSecret.Data[v1.ServiceAccountTokenKey])
	}
	return
}

// getSecret returns the secret, taking the namespace from GitRepository if it is empty
func (c *ClientFactory) getSecret(ref *v1.SecretReference) (secret *v1.Secret, err error) {
	secret = &v1.Secret{}
	ns := ref.Namespace

	if err = c.k8sClient.Get(context.TODO(), types.NamespacedName{
		Namespace: ns, Name: ref.Name,
	}, secret); err != nil {
		err = fmt.Errorf("cannot get secret %v, error is: %v", secret, err)
	}
	return
}

// ResourceGetter represent the interface for getting Kubernetes resource
type ResourceGetter interface {
	Get(ctx context.Context, key types.NamespacedName, obj runtime.Object) error
}
