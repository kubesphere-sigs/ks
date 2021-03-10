package auth

import "fmt"

func getGitHubAuth(opt authAddOption) string {
	return fmt.Sprintf(`
name: GitHub
type: GitHubIdentityProvider
mappingMethod: auto
provider:
  clientID: %s
  clientSecret: %s
  endpoint:
    authURL: 'https://github.com/login/oauth/authorize'
    tokenURL: 'https://github.com/login/oauth/access_token'
  redirectURL: %s/auth/redirect
  scopes:
  - user
`, opt.ClientID, opt.ClientSecret, opt.Host)
}
