package app

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"reflect"
	"testing"
)

func Test_getApplication(t *testing.T) {
	defaultApp := &unstructured.Unstructured{}
	defaultApp.SetKind("Application")
	defaultApp.SetAPIVersion("gitops.kubesphere.io/v1alpha1")
	defaultApp.SetNamespace("ns")
	defaultApp.SetName("fake")
	_ = unstructured.SetNestedField(defaultApp.Object, "http://git.com", "spec", "argoApp", "source", "repoURL")
	_ = unstructured.SetNestedField(defaultApp.Object, "master", "spec", "argoApp", "source", "targetRevision")
	_ = unstructured.SetNestedField(defaultApp.Object, "current", "spec", "argoApp", "source", "path")

	invalidApp := defaultApp.DeepCopy()
	_ = unstructured.SetNestedField(invalidApp.Object, "", "spec", "argoApp", "source", "repoURL")

	scheme := runtime.NewScheme()

	type args struct {
		name      string
		namespace string
		client    dynamic.Interface
	}
	tests := []struct {
		name    string
		args    args
		wantApp *application
		wantErr bool
	}{{
		name: "normal case",
		args: args{
			name:      "fake",
			namespace: "ns",
			client:    fake.NewSimpleDynamicClient(scheme, defaultApp.DeepCopy()),
		},
		wantApp: &application{
			namespace: "ns",
			name:      "fake",
			gitRepo:   "http://git.com",
			branch:    "master",
			directory: "current",
		},
		wantErr: false,
	}, {
		name: "invalid application",
		args: args{
			name:      "fake",
			namespace: "ns",
			client:    fake.NewSimpleDynamicClient(scheme, invalidApp.DeepCopy()),
		},
		wantErr: true,
	}, {
		name: "not found application",
		args: args{
			name:      "fake",
			namespace: "ns",
			client:    fake.NewSimpleDynamicClient(scheme),
		},
		wantErr: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotApp, err := getApplication(tt.args.name, tt.args.namespace, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("getApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotApp, tt.wantApp) {
				t.Errorf("getApplication() gotApp = %v, want %v", gotApp, tt.wantApp)
			}
		})
	}
}
