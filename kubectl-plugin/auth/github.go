package auth

import "fmt"

var e = `
name: GitHub
type: GitHubIdentityProvider
mappingMethod: auto
provider:
  clientID: 76b2f45277bb5314457f
  clientSecret: ed04cf65d99cb7818a6eb11a72b77efcedef9c24
  redirectURL: "http://139.198.3.176:30880/oauth/redirect"
  scopes:
  - user
`

func getGitHubAuth(opt authOption) string {
	return fmt.Sprintf(`
name: GitHub
type: GitHubIdentityProvider
mappingMethod: auto
provider:
  clientID: %s
  clientSecret: %s
  redirectURL: "%s"
  scopes:
  - user
`, opt.ClientID, opt.ClientSecret, opt.RedirectURL)
}
