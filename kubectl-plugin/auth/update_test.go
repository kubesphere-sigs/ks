package auth

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestUpdate(t *testing.T) {
	result := updateAuthentication(yamlFile, "GitHub", getGitHubAuth(authAddOption{
		ClientID:     "76b2f45277bb5314457f",
		ClientSecret: "ed04cf65d99cb7818a6eb11a72b77efcedef9c24",
		Host:         "http://139.198.3.176:30880",
	}))

	expectedMap := make(map[string]interface{}, 0)
	var expectedData []byte
	err := yaml.Unmarshal([]byte(expected), expectedMap)
	assert.Nil(t, err)
	expectedData, err = yaml.Marshal(expectedMap)
	assert.Nil(t, err)
	assert.Equal(t, result, string(expectedData))
}

func TestUpdateWhichHasAuth(t *testing.T) {
	result := updateAuthentication(hasAuthYaml, "GitHub", "e")

	expectedMap := make(map[string]interface{}, 0)
	var expectedData []byte
	err := yaml.Unmarshal([]byte(expectedhasAuthYaml), expectedMap)
	assert.Nil(t, err)
	expectedData, err = yaml.Marshal(expectedMap)
	assert.Nil(t, err)
	assert.Equal(t, result, string(expectedData))
}

var expected = `authentication:
    authenticateRateLimiterDuration: 10m0s
    authenticateRateLimiterMaxTries: 10
    jwtSecret: Fa9bYHL24E1u7tMwHM9Ko1iwQ77aTRXE
    kubectlImage: kubespheredev/kubectl:v1.17.0
    loginHistoryRetentionPeriod: 168h
    maximumClockSkew: 10s
    multipleLogin: true
    oauthOptions:
        accessTokenInactivityTimeout: 30m
        accessTokenMaxAge: 1h
        identityProviders:
            - mappingMethod: auto
              name: GitHub
              provider:
                clientID: 76b2f45277bb5314457f
                clientSecret: ed04cf65d99cb7818a6eb11a72b77efcedef9c24
                endpoint:
                    authURL: https://github.com/login/oauth/authorize
                    tokenURL: https://github.com/login/oauth/access_token
                redirectURL: http://139.198.3.176:30880/auth/redirect
                scopes:
                    - user
              type: GitHubIdentityProvider
ldap:
    groupSearchBase: ou=Groups,dc=kubesphere,dc=io
    host: openldap.kubesphere-system.svc:389
    managerDN: cn=admin,dc=kubesphere,dc=io
    managerPassword: admin
    userSearchBase: ou=Users,dc=kubesphere,dc=io
redis:
    db: 0
    host: redis.kubesphere-system.svc
    password: ""
    port: 6379
`

var yamlFile = `
authentication:
  authenticateRateLimiterMaxTries: 10
  authenticateRateLimiterDuration: 10m0s
  loginHistoryRetentionPeriod: 168h
  maximumClockSkew: 10s
  multipleLogin: True
  kubectlImage: kubespheredev/kubectl:v1.17.0
  jwtSecret: "Fa9bYHL24E1u7tMwHM9Ko1iwQ77aTRXE"
ldap:
  host: openldap.kubesphere-system.svc:389
  managerDN: cn=admin,dc=kubesphere,dc=io
  managerPassword: admin
  userSearchBase: ou=Users,dc=kubesphere,dc=io
  groupSearchBase: ou=Groups,dc=kubesphere,dc=io
redis:
  host: redis.kubesphere-system.svc
  port: 6379
  password: ""
  db: 0
`

var expectedhasAuthYaml = `authentication:
    authenticateRateLimiterDuration: 10m0s
    authenticateRateLimiterMaxTries: 10
    identityProviders:
        - mappingMethod: auto
          name: fake
    jwtSecret: Fa9bYHL24E1u7tMwHM9Ko1iwQ77aTRXE
    kubectlImage: kubespheredev/kubectl:v1.17.0
    loginHistoryRetentionPeriod: 168h
    maximumClockSkew: 10s
    multipleLogin: true
    oauthOptions:
        accessTokenInactivityTimeout: 30m
        accessTokenMaxAge: 1h
        identityProviders:
            - {}
ldap:
    groupSearchBase: ou=Groups,dc=kubesphere,dc=io
    host: openldap.kubesphere-system.svc:389
    managerDN: cn=admin,dc=kubesphere,dc=io
    managerPassword: admin
    userSearchBase: ou=Users,dc=kubesphere,dc=io
redis:
    db: 0
    host: redis.kubesphere-system.svc
    password: ""
    port: 6379
`

var hasAuthYaml = `
authentication:
  authenticateRateLimiterMaxTries: 10
  authenticateRateLimiterDuration: 10m0s
  loginHistoryRetentionPeriod: 168h
  maximumClockSkew: 10s
  multipleLogin: True
  kubectlImage: kubespheredev/kubectl:v1.17.0
  jwtSecret: "Fa9bYHL24E1u7tMwHM9Ko1iwQ77aTRXE"
  identityProviders:
  - mappingMethod: auto
    name: fake
ldap:
  host: openldap.kubesphere-system.svc:389
  managerDN: cn=admin,dc=kubesphere,dc=io
  managerPassword: admin
  userSearchBase: ou=Users,dc=kubesphere,dc=io
  groupSearchBase: ou=Groups,dc=kubesphere,dc=io
redis:
  host: redis.kubesphere-system.svc
  port: 6379
  password: ""
  db: 0
`
