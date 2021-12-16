package auth

import "fmt"

func getGiteeAuth(opt authAddOption) string {
	return fmt.Sprintf(`
name: Gitee
type: GitHubIdentityProvider
mappingMethod: auto
provider:
  clientID: %s
  clientSecret: %s
  redirectURL: "%s"
  endpoint:
    authURL: 'https://gitee.com/oauth/authorize'
    tokenURL: 'https://gitee.com/oauth/token'
  redirectURL: %s/auth/redirect
  scopes:
  - user_info
`, opt.ClientID, opt.ClientSecret, opt.Host, opt.Host)
}
