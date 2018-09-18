package environment_test

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/fabric8-services/fabric8-tenant/environment"
	"regexp"
	"github.com/fabric8-services/fabric8-tenant/test/doubles"
	"sort"
)

var processTemplate = `
- apiVersion: v1
  kind: Project
  metadata:
    annotations:
      openshift.io/description: ${PROJECT_DESCRIPTION}
      openshift.io/display-name: ${PROJECT_DISPLAYNAME}
      openshift.io/requester: ${PROJECT_REQUESTING_USER}
      serviceaccounts.openshift.io/oauth-redirectreference.jenkins: '{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"jenkins"}}'
    labels:
      provider: fabric8
      project: fabric8-tenant-team-environments
      version: 1.0.58
      group: io.fabric8.online.packages
    name: ${PROJECT_NAME}
    credentials.xml.tpl: |-
      <?xml version='1.0' encoding='UTF-8'?>
      <com.cloudbees.plugins.credentials.SystemCredentialsProvider plugin="credentials@1.23">
      </com.cloudbees.plugins.credentials.SystemCredentialsProvider>
`

var parseNamespaceTemplate = `
---
apiVersion: v1
kind: Template
metadata:
  labels:
    provider: fabric8
    project: fabric8-tenant-team-environments
    version: 1.0.58
    group: io.fabric8.online.packages
  name: fabric8-tenant-team-envi
objects:
- apiVersion: v1
  kind: Namespace
  metadata:
    annotations:
      openshift.io/description: Test-Project-Description
      openshift.io/display-name: Test-Project-Name
      openshift.io/requester: Aslak-User
    labels:
      provider: fabric8
      project: fabric8-tenant-team-environments
      version: 1.0.58
      group: io.fabric8.online.packages
    name: aslak-test
- apiVersion: v1
  kind: RoleBindingRestriction
  metadata:
    labels:
      app: fabric8-tenant-che-mt
      provider: fabric8
      version: 2.0.85
      group: io.fabric8.tenant.packages
    name: dsaas-user-access
  spec:
    userrestriction:
      users:
      - ${PROJECT_USER}
`
var processTemplateVariables = `
- apiVersion: v1
  kind: Project
  metadata:
    labels:
      provider: fabric8
      project: fabric8-tenant-team-environments
      version: 1.0.58
      group: io.fabric8.online.packages
    credentials.xml.tpl: |-
      <?xml version='1.0' encoding='UTF-8'?>
      <com.cloudbees.plugins.credentials.SystemCredentialsProvider plugin="credentials@1.23">
        <domainCredentialsMap class="hudson.util.CopyOnWriteMap$Hash">
          ${KUBERNETES_CREDENTIALS}
        </domainCredentialsMap>
      </com.cloudbees.plugins.credentials.SystemCredentialsProvider>
`
var sortTemplate = `
---
apiVersion: v1
kind: Template
objects:
- apiVersion: v1
  kind: Secret
  metadata:
    name: aslak-test
- apiVersion: v1
  kind: ProjectRequest
  metadata:
    name: aslak-test
- apiVersion: v1
  kind: ServiceAccount
  metadata:
    name: aslak-test
- apiVersion: v1
  kind: RoleBinding
  metadata:
    name: aslak-test
- apiVersion: v1
  kind: RoleBindingRestriction
  metadata:
    name: aslak-test
- apiVersion: v1
  kind: ResourceQuota
  metadata:
    name: aslak-test
- apiVersion: v1
  kind: LimitRange
  metadata:
    name: aslak-test
`

func TestSort(t *testing.T) {
	data := testdoubles.LoadTestConfig(t)
	template := environment.Template{Content:sortTemplate}
	objects, err := template.Process(environment.CollectVars("developer", data))
	sort.Sort(environment.ByKind(objects))
	require.NoError(t, err)

	assert.Equal(t, "ProjectRequest", kind(objects[0]))
	assert.Equal(t, "RoleBindingRestriction", kind(objects[1]))
	assert.Equal(t, "LimitRange", kind(objects[2]))
	assert.Equal(t, "ResourceQuota", kind(objects[3]))
}

func TestParseNamespace(t *testing.T) {
	template := environment.Template{Content:parseNamespaceTemplate}
	objects, err := template.Process(environment.CollectVars("developer", testdoubles.LoadTestConfig(t)))
	require.NoError(t, err)

	assert.Equal(t, "Namespace", kind(objects[0]))
	assert.Equal(t, "RoleBindingRestriction", kind(objects[1]))
}

func kind(object environment.Object) string {
	return object["kind"].(string)
}

var dnsRegExp = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"

func TestCreateUsername(t *testing.T) {
	assertName(t, "some", "some@email.com")
	assertName(t, "so-me", "so-me@email.com")
	assertName(t, "some", "some")
	assertName(t, "so-me", "so-me")
	assertName(t, "so-me", "so_me")
	assertName(t, "so-me", "so me")
	assertName(t, "so-me", "so me@email.com")
	assertName(t, "so-me", "so.me")
	assertName(t, "so-me", "so?me")
	assertName(t, "so-me", "so:me")
	assertName(t, "some1", "some1")
	assertName(t, "so1me1", "so1me1")
}

func assertName(t *testing.T, expected, username string) {
	assert.Regexp(t, dnsRegExp, environment.RetrieveUserName(username))
	assert.Equal(t, expected, environment.RetrieveUserName(username))
}

func TestProcess(t *testing.T) {
	vars := map[string]string{
		"PROJECT_DESCRIPTION":     "Test-Project-Description",
		"PROJECT_DISPLAYNAME":     "Test-Project-Name",
		"PROJECT_REQUESTING_USER": "Aslak-User",
		"PROJECT_NAME":            "Aslak-Test",
	}
	template := environment.Template{Content:processTemplate}
	processed, err := template.ReplaceVars(vars)
	require.Nil(t, err, "error processing templateDef")

	t.Run("verify no templateDef markers in output", func(t *testing.T) {
		assert.False(t, regexp.MustCompile(`\${([A-Z_]+)}`).MatchString(processed))
	})
	t.Run("verify markers were replaced", func(t *testing.T) {
		assert.Contains(t, processed, vars["PROJECT_DESCRIPTION"], "missing")
		assert.Contains(t, processed, vars["PROJECT_DISPLAYNAME"], "missing")
		assert.Contains(t, processed, vars["PROJECT_REQUESTING_USER"], "missing")
		assert.Contains(t, processed, vars["PROJECT_NAME"], "missing")
	})
	t.Run("Verify not fiddling with values", func(t *testing.T) {
		assert.Contains(t, processed, `'{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"jenkins"}}'`)
	})

	t.Run("Verify not escaping xml/html values", func(t *testing.T) {
		assert.Contains(t, processed, `<?xml version='1.0' encoding='UTF-8'?>`)
	})
}

func TestProcessVariables(t *testing.T) {
	vars := map[string]string{}

	template := environment.Template{Content:processTemplateVariables}
	processed, err := template.ReplaceVars(vars)
	require.Nil(t, err, "error processing templateDef")

	t.Run("Verify non replaced markers are left", func(t *testing.T) {
		assert.Contains(t, processed, "${KUBERNETES_CREDENTIALS}", "missing")
	})
}
