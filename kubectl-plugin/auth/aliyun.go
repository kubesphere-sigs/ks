package auth

import "fmt"

var a = `
      oauthOptions:
        identityProviders:
        - name: AliyunIDaas
          type: AliyunIDaaSProvider
          mappingMethod: auto
          provider:
            clientID: a8a0ad386ae2c12133e33f308c8fc4f0nP1wmb7PHfl
            clientSecret: sQRYYJ6Xe8ZUEvPCsthFN6HznCqAd97bdYBA3JWCAJ
            grantType: authorization_code
            endpoint:
              userInfoUrl: "https://huedxurbjj.login.aliyunidaas.com/api/bff/v1.2/oauth2/userinfo"
              authURL: "https://huedxurbjj.login.aliyunidaas.com/oauth/authorize"
              tokenURL: "https://huedxurbjj.login.aliyunidaas.com/oauth/token"
            redirectURL: "http://139.198.3.176:30880/oauth/redirect"
            scopes:
            - read
`

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

var b = `
subjects:
- apiGroup: iam.kubesphere.io/v1alpha2
  kind: User
  name: admin
- apiGroup: users.iam.kubesphere.io
  kind: Group
  name: pre-registration
`
