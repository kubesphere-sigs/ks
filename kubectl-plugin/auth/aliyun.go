package auth

import "fmt"

func getAliyunAuth(opt authOption) string {
	return fmt.Sprintf(`
name: Aliyun
type: AliyunIDaaSProvider
mappingMethod: auto
provider:
  clientID: %s
  clientSecret: %s
  grantType: authorization_code
  endpoint:
    userInfoUrl: "https://huedxurbjj.login.aliyunidaas.com/api/bff/v1.2/oauth2/userinfo"
    authURL: "https://huedxurbjj.login.aliyunidaas.com/oauth/authorize"
    tokenURL: "https://huedxurbjj.login.aliyunidaas.com/oauth/token"
  redirectURL: "%s"
  scopes:
  - read
`, opt.ClientID, opt.ClientSecret, opt.RedirectURL)
}
