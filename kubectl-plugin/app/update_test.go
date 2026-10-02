package app

import (
	"os"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

func Test_getApplication(t *testing.T) {
	defaultApp := &unstructured.Unstructured{}
	defaultApp.SetKind("Application")
	defaultApp.SetAPIVersion("gitops.kubesphere.io/v1alpha1")
	defaultApp.SetNamespace("ns")
	defaultApp.SetName("fake")
	_ = unstructured.SetNestedField(defaultApp.Object, "http://git.com", "spec", "argoApp", "spec", "source", "repoURL")
	_ = unstructured.SetNestedField(defaultApp.Object, "master", "spec", "argoApp", "spec", "source", "targetRevision")
	_ = unstructured.SetNestedField(defaultApp.Object, "current", "spec", "argoApp", "spec", "source", "path")

	invalidApp := defaultApp.DeepCopy()
	_ = unstructured.SetNestedField(invalidApp.Object, "", "spec", "argoApp", "spec", "source", "repoURL")

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

func Test_resolveTargetRevision(t *testing.T) {
	hash := plumbing.NewHash("8b3f4a51e5a5d0c84b1f6b0a3c7c8f5f2e0f7d61")
	refs := []*plumbing.Reference{
		plumbing.NewHashReference(plumbing.NewBranchReferenceName("master"), hash),
		plumbing.NewHashReference(plumbing.NewBranchReferenceName("dev"), hash),
		plumbing.NewHashReference(plumbing.NewTagReferenceName("v1.0.0"), hash),
	}

	tests := []struct {
		name         string
		revision     string
		wantRefName  plumbing.ReferenceName
		wantIsCommit bool
		wantErr      bool
	}{{
		name:     "empty revision uses the default branch",
		revision: "",
	}, {
		name:     "HEAD uses the default branch",
		revision: "HEAD",
	}, {
		name:        "short branch name",
		revision:    "dev",
		wantRefName: "refs/heads/dev",
	}, {
		name:        "short tag name",
		revision:    "v1.0.0",
		wantRefName: "refs/tags/v1.0.0",
	}, {
		name:        "full reference name",
		revision:    "refs/heads/master",
		wantRefName: "refs/heads/master",
	}, {
		name:         "full commit SHA",
		revision:     hash.String(),
		wantIsCommit: true,
	}, {
		name:         "abbreviated commit SHA",
		revision:     "8b3f4a5",
		wantIsCommit: true,
	}, {
		name:     "unknown branch",
		revision: "not-exist",
		wantErr:  true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRefName, gotIsCommit, err := resolveTargetRevision(tt.revision, refs)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveTargetRevision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRefName != tt.wantRefName {
				t.Errorf("resolveTargetRevision() refName = %v, want %v", gotRefName, tt.wantRefName)
			}
			if gotIsCommit != tt.wantIsCommit {
				t.Errorf("resolveTargetRevision() isCommit = %v, want %v", gotIsCommit, tt.wantIsCommit)
			}
		})
	}
}

func Test_cloneGitRepo(t *testing.T) {
	// prepare a repository whose default branch is master, and has a dev branch and a v1.0.0 tag
	remoteDir := t.TempDir()
	remote, err := git.PlainInit(remoteDir, false)
	if err != nil {
		t.Fatal(err)
	}
	wd, err := remote.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	commit := func(file string) plumbing.Hash {
		if err := os.WriteFile(path.Join(remoteDir, file), []byte(file), 0644); err != nil {
			t.Fatal(err)
		}
		if _, err := wd.Add(file); err != nil {
			t.Fatal(err)
		}
		hash, err := wd.Commit(file, &git.CommitOptions{
			Author: &object.Signature{Name: "test", Email: "test@example.com", When: time.Now()},
		})
		if err != nil {
			t.Fatal(err)
		}
		return hash
	}
	masterCommit := commit("master.txt")
	if _, err = remote.CreateTag("v1.0.0", masterCommit, nil); err != nil {
		t.Fatal(err)
	}
	if err = wd.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("dev"), Create: true}); err != nil {
		t.Fatal(err)
	}
	devCommit := commit("dev.txt")
	if err = wd.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("master")}); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		revision   string
		wantCommit plumbing.Hash
		wantErr    bool
	}{
		{name: "HEAD", revision: "HEAD", wantCommit: masterCommit},
		{name: "short branch name", revision: "dev", wantCommit: devCommit},
		{name: "full reference name", revision: "refs/heads/dev", wantCommit: devCommit},
		{name: "short tag name", revision: "v1.0.0", wantCommit: masterCommit},
		{name: "commit SHA", revision: devCommit.String(), wantCommit: devCommit},
		{name: "abbreviated commit SHA", revision: devCommit.String()[:7], wantCommit: devCommit},
		{name: "unknown branch", revision: "not-exist", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := cloneGitRepo(&application{gitRepo: remoteDir, branch: tt.revision}, nil)
			defer func() {
				_ = os.RemoveAll(dir)
			}()
			if (err != nil) != tt.wantErr {
				t.Fatalf("cloneGitRepo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			repo, err := git.PlainOpen(dir)
			if err != nil {
				t.Fatal(err)
			}
			head, err := repo.Head()
			if err != nil {
				t.Fatal(err)
			}
			if head.Hash() != tt.wantCommit {
				t.Errorf("cloneGitRepo() HEAD = %v, want %v", head.Hash(), tt.wantCommit)
			}
		})
	}
}
