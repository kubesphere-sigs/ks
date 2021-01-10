package auth

import "fmt"

func getGitHubAuth(opt authOption) string {
	return fmt.Sprintf(`
name: GitHub
type: GitHubIdentityProvider
mappingMethod: auto
provider:
  clientID: %s
  clientSecret: %s
  redirectURL: "%s"
  endpoint:
    authURL: 'https://github.com/login/oauth/authorize'
    tokenURL: 'https://github.com/login/oauth/access_token'
  scopes:
  - user
`, opt.ClientID, opt.ClientSecret, opt.RedirectURL)
}
