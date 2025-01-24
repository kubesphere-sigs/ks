package app

import (
	"context"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/common"
	"github.com/kubesphere-sigs/ks/kubectl-plugin/types"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

func newUpdateCmd(client dynamic.Interface, clientset *kubernetes.Clientset) (cmd *cobra.Command) {
	opt := updateOption{client: client, clientset: clientset}
	cmd = &cobra.Command{
		Use:     "update",
		Aliases: []string{"up"},
		Short:   "Update the files which in the git repository. Support kustomize only.",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}

	flags := cmd.Flags()
	flags.StringVarP(&opt.appName, "app-name", "", "",
		"The name of the application")
	flags.StringVarP(&opt.appNamespace, "app-namespace", "", "",
		"The namespace of the application")
	flags.StringVarP(&opt.name, "name", "n", "",
		"Name is a tag-less image name")
	flags.StringVarP(&opt.newName, "newName", "", "",
		"NewName is the value used to replace the original name")
	flags.StringVarP(&opt.newTag, "newTag", "t", "",
		"NewTag is the value used to replace the original tag")
	flags.StringVarP(&opt.digest, "digest", "d", "",
		"Digest is the value used to replace the original image tag. If digest is present NewTag value is ignored")
	flags.StringVarP(&opt.mode, "mode", "m", "commit",
		"The way you want to push the changes to the target git repository, support mode: pr, commit")
	flags.StringVarP(&opt.gitProvider, "git-provider", "", "",
		"The flag --mode=pr need the git provider, the mode will fallback to commit if the git provider is empty")
	flags.StringVarP(&opt.gitPassword, "git-password", "", "",
		"The password of the git provider")
	flags.StringVarP(&opt.gitUsername, "git-username", "", "",
		"The username of the git provider")
	flags.StringVarP(&opt.gitEmail, "git-email", "", "",
		"The email of the git provider")
	flags.StringVarP(&opt.gitTargetBranch, "git-target-branch", "", "",
		"The target branch name that you want to push")
	flags.StringVarP(&opt.secretName, "secret-name", "", "",
		"The username of the git provider")
	flags.StringVarP(&opt.secretNamespace, "secret-namespace", "", "",
		"The username of the git provider")

	_ = cmd.MarkFlagRequired("app-name")
	_ = cmd.MarkFlagRequired("app-namespace")
	_ = cmd.Flags().MarkHidden("mode")
	return
}

func (o *updateOption) preRunE(_ *cobra.Command, _ []string) (err error) {
	switch o.mode {
	case "pr", "commit":
	default:
		err = fmt.Errorf("supportted value: pr, commit. Please check the flag --mode")
		return
	}
	return
}

func (o *updateOption) runE(cmd *cobra.Command, args []string) (err error) {
	// find the app cr
	var app *application
	if app, err = getApplication(o.appName, o.appNamespace, o.client); err != nil {
		err = fmt.Errorf("cannot find application '%s/%s', error is: %v", o.appNamespace, o.appName, err)
		return
	}

	var gitAuth transport.AuthMethod
	if gitAuth, err = o.getGitAuth(); err != nil {
		err = fmt.Errorf("failed to create git auth, error is: %v", err)
		return
	}

	// clone the target git repo
	var tempDir string
	if tempDir, err = cloneGitRepo(app, gitAuth); err != nil {
		err = fmt.Errorf("failed to clone git repository '%s', error is: %v", app.gitRepo, err)
		return
	}
	fmt.Println("git-repo-dir: ", tempDir)

	// run kustomize command
	if err = updateKustomization(o, path.Join(tempDir, app.directory)); err != nil {
		err = fmt.Errorf("failed to update the kustomization, error is: %v", err)
		return
	}

	// git commit the changes
	repo := getForkAppRepo(app)
	repo.localGitRepoDir = tempDir
	repo.token = o.gitPassword
	repo.username = o.gitUsername
	repo.email = o.gitEmail
	repo.gitAuth = gitAuth
	repo.gitTargetBranch = o.gitTargetBranch
	if o.gitUsername != "" {
		repo.botOrg = o.gitUsername
	}
	if err = pushChanges(o.mode, repo); err != nil {
		err = fmt.Errorf("failed to push changes, error is: %v", err)
	}
	return
}

func (o *updateOption) getGitAuth() (auth transport.AuthMethod, err error) {
	if o.gitUsername != "" && o.gitPassword != "" {
		auth = getGitUsernameAndPasswordAuth(o.gitUsername, o.gitPassword)
	} else if o.secretNamespace != "" && o.secretName != "" {
		var secret *v1.Secret
		if secret, err = o.clientset.CoreV1().Secrets(o.secretNamespace).
			Get(context.Background(), o.secretName, metav1.GetOptions{}); err != nil {
			return
		}

		auth = getAuth(secret)
	}
	return
}

func getForkAppRepo(app *application) *forkAppRepo {
	var gitProvider string
	var org string
	var repo string

	if strings.Contains(app.gitRepo, "github.com") {
		gitProvider = "github"
	} else if strings.Contains(app.gitRepo, "gitlab.com") {
		gitProvider = "gitlab"
	} else if strings.Contains(app.gitRepo, "gitee.com") {
		gitProvider = "gitee"
	} else if strings.Contains(app.gitRepo, "gitea.com") {
		gitProvider = "gitea"
	} else if strings.Contains(app.gitRepo, "bitbucket.com") {
		gitProvider = "bitbucket"
	}

	if gitProvider != "" {
		orgAndRepo := strings.ReplaceAll(app.gitRepo, fmt.Sprintf("https://%s.com/", gitProvider), "")
		orgAndRepo = strings.TrimSuffix(orgAndRepo, "/")
		org = strings.Split(orgAndRepo, "/")[0]
		repo = strings.Split(orgAndRepo, "/")[1]
	}

	return &forkAppRepo{
		application: app,
		gitProvider: gitProvider,
		org:         org,
		repo:        repo,
		forkRemote:  "bot",            // TODO change it later
		botOrg:      "linuxsuren-bot", // TODO change it later
	}
}

func getApplication(name, namespace string, client dynamic.Interface) (app *application, err error) {
	var unstructedObject *unstructured.Unstructured
	if unstructedObject, err = client.Resource(types.GetApplicationSchema()).Namespace(namespace).
		Get(context.Background(), name, metav1.GetOptions{}); err != nil {
		return
	}

	gitRepo, _, _ := unstructured.NestedString(unstructedObject.Object, "spec", "argoApp", "spec", "source", "repoURL")
	branch, _, _ := unstructured.NestedString(unstructedObject.Object, "spec", "argoApp", "spec", "source", "targetRevision")
	directory, _, _ := unstructured.NestedString(unstructedObject.Object, "spec", "argoApp", "spec", "source", "path")

	app = &application{
		namespace: namespace,
		name:      name,
		gitRepo:   gitRepo,
		branch:    branch,
		directory: directory,
	}
	if app.gitRepo == "" || app.branch == "" || app.directory == "" {
		app = nil
		err = fmt.Errorf("only support git repository as the source")
	}
	return
}

// cloneGitRepo clones the git repository with a generic git client
func cloneGitRepo(app *application, auth transport.AuthMethod) (tempDir string, err error) {
	if tempDir, err = os.MkdirTemp(os.TempDir(), ""); err != nil {
		err = fmt.Errorf("failed to create a temp directory, error is %v", err)
		return
	}

	if _, err = git.PlainClone(tempDir, false, &git.CloneOptions{
		URL:           app.gitRepo,
		ReferenceName: plumbing.ReferenceName(app.branch),
		Auth:          auth,
	}); err != nil {
		err = fmt.Errorf("failed to clone git repository '%s' into '%s', error: %v",
			app.gitRepo, tempDir, err)
	}
	return
}

func updateKustomization(patch *updateOption, dir string) (err error) {
	var digest string
	if patch.digest != "" {
		digest = fmt.Sprintf("@%s", patch.digest)
	}
	var tag string
	if patch.newTag != "" {
		tag = fmt.Sprintf(":%s", patch.newTag)
	}

	imagePatch := fmt.Sprintf("%s=%s%s%s", patch.name, patch.newName, tag, digest)
	_ = common.ExecCommandInDirectory("kustomize", dir, "edit", "set", "image", imagePatch)
	return
}

func pushToBranch(forkApp *forkAppRepo, mode string) (err error) {
	// add a bot account
	var repo *git.Repository
	if repo, err = git.PlainOpen(forkApp.localGitRepoDir); err == nil {
		var wd *git.Worktree

		if wd, err = repo.Worktree(); err == nil {
			if mode == "pr" {
				if err = makeSureRemote(forkApp.forkRemote, forkApp.forkRepoAddress, repo); err != nil {
					return
				}

				if err = wd.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName(forkApp.forkBranch),
					Create: false,
					Keep:   true,
				}); err != nil {
					err = fmt.Errorf("unable to checkout git branch: %s, error: %v", forkApp.forkBranch, err)
					return
				}
			} else if forkApp.gitTargetBranch != "" {
				var remoteBranches []string
				remoteBranches, err = listRemoteBranches(repo, forkApp.gitAuth)
				if err != nil {
					err = fmt.Errorf("failed to list remote branches, error is: %+v", err)
					return
				}
				create := true
				if slices.Contains(remoteBranches, forkApp.gitTargetBranch) {
					create = false
				}
				if err = wd.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName(forkApp.gitTargetBranch),
					Create: create,
					Keep:   true,
				}); err != nil {
					err = fmt.Errorf("unable to checkout git branch: %s, error: %v", forkApp.forkBranch, err)
					return
				}
			}

			// commit and push
			if _, err = wd.Add("."); err != nil {
				err = fmt.Errorf("failed to run git add command, error is %v", err)
				return
			}
			if _, err = wd.Commit("files changed by ks cli", &git.CommitOptions{
				Author: &object.Signature{
					Name:  forkApp.username,
					Email: forkApp.email,
					When:  time.Now(),
				},
			}); err != nil {
				err = fmt.Errorf("failed to commit changes, error is %v", err)
				return
			}
			err = repo.Push(&git.PushOptions{
				Auth: forkApp.gitAuth,
			})
		}
	}
	return
}

func listRemoteBranches(repo *git.Repository, auth transport.AuthMethod) ([]string, error) {
	remote, err := repo.Remote("origin")
	if err != nil {
		return nil, err
	}
	refs, err := remote.List(&git.ListOptions{
		Auth:            auth,
		InsecureSkipTLS: true,
	})
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, ref := range refs {
		branches = append(branches, ref.Name().Short())
	}
	return branches, nil
}

func makeSureRemote(name, repoAddr string, repo *git.Repository) (err error) {
	if _, err = repo.Remote(name); err != nil {
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: name,
			URLs: []string{repoAddr},
		})
	}
	return
}

func getGitAuthAsToken(token string) transport.AuthMethod {
	secret := &v1.Secret{
		Type: v1.SecretTypeOpaque,
		Data: map[string][]byte{
			v1.ServiceAccountTokenKey: []byte(token),
		},
	}
	return getAuth(secret)
}

func getGitUsernameAndPasswordAuth(username, password string) transport.AuthMethod {
	secret := &v1.Secret{
		Type: v1.SecretTypeBasicAuth,
		Data: map[string][]byte{
			v1.BasicAuthUsernameKey: []byte(username),
			v1.BasicAuthPasswordKey: []byte(password),
		},
	}
	return getAuth(secret)
}

func getAuth(secret *v1.Secret) (auth transport.AuthMethod) {
	if secret == nil {
		return
	}

	switch secret.Type {
	case v1.SecretTypeOpaque:
		auth = &githttp.TokenAuth{
			Token: string(secret.Data[v1.ServiceAccountTokenKey]),
		}
	case v1.SecretTypeBasicAuth:
		auth = &githttp.BasicAuth{
			Username: string(secret.Data[v1.BasicAuthUsernameKey]),
			Password: string(secret.Data[v1.BasicAuthPasswordKey]),
		}
	case v1.SecretTypeSSHAuth:
		signer, _ := ssh.ParsePrivateKey(secret.Data[v1.SSHAuthPrivateKey])
		auth = &gitssh.PublicKeys{User: "git", Signer: signer}
	}
	return
}

func pushChanges(mode string, forkApp *forkAppRepo) (err error) {
	provider := forkApp.gitProvider
	repo := forkApp.repo
	ns := forkApp.org
	botNs := forkApp.botOrg
	orgAndRepo := fmt.Sprintf("%s/%s", ns, repo)
	branchName := fmt.Sprintf("%d", time.Now().Nanosecond())
	forkApp.forkBranch = branchName

	if mode == "pr" {
		var client *scm.Client
		if client, err = NewClientFactory(provider, forkApp.token, nil, nil).GetClient(); err == nil && client != nil {
			in := &scm.RepositoryInput{
				Namespace: botNs,
			}
			var forkedRepo *scm.Repository
			if forkedRepo, _, err = client.Repositories.Fork(context.Background(), in, orgAndRepo); err != nil {
				err = fmt.Errorf("failed to fork repo: %s, error is: %v", orgAndRepo, err)
				return
			}

			forkApp.forkRepoAddress = forkedRepo.Clone
			if err = pushToBranch(forkApp, mode); err != nil {
				err = fmt.Errorf("failed to push branch, error is: %v", err)
				return
			}

			input := &scm.PullRequestInput{
				Title: "Amazing new feature",
				Body:  "Please pull these awesome changes in!",
				Head:  fmt.Sprintf("%s:%s", botNs, branchName),
				Base:  "master",
			}

			_, _, err = client.PullRequests.Create(context.Background(), orgAndRepo, input)
		}
	} else {
		if err = pushToBranch(forkApp, mode); err != nil {
			err = fmt.Errorf("failed to push branch, error is: %v", err)
			return
		}
	}
	return
}

type forkAppRepo struct {
	*application

	gitProvider     string
	gitTargetBranch string
	org, repo       string
	botOrg          string
	localGitRepoDir string
	forkRemote      string
	forkRepoAddress string
	forkBranch      string
	token           string
	username        string
	email           string
	gitAuth         transport.AuthMethod
}

type application struct {
	namespace string
	name      string
	gitRepo   string
	branch    string
	directory string
}

type updateOption struct {
	appName      string
	appNamespace string
	name         string
	newName      string
	newTag       string
	digest       string
	// mode indicates how to update the application git repository, pr or commit
	mode            string
	gitProvider     string
	gitPassword     string
	gitUsername     string
	gitEmail        string
	gitTargetBranch string

	// secretName and secretNamespace are used to the git auth
	secretName      string
	secretNamespace string

	client    dynamic.Interface
	clientset *kubernetes.Clientset
}
