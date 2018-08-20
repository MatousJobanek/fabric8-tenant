package openshift_test

import (
	"testing"
	"gopkg.in/h2non/gock.v1"
	"os"
	"github.com/fabric8-services/fabric8-tenant/openshift"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	"github.com/fabric8-services/fabric8-tenant/test/doubles"
)

var templateHeader = `
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
`
var namespaceObject = `
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
`
var roleBindingRestrictionObject = `
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

var openshiftClient *openshift.ServiceBuilder

func TestMain(m *testing.M) {
	defer gock.Off()
	retCode := m.Run()
	os.Exit(retCode)
}

func createOpenshiftClient(config *configuration.Data) *openshift.ServiceBuilder {
	return testdoubles.NewOpenshiftService("http://starter.com", "USFpK7R-YBRlRONI5Ru-GakBtP7fr891rg", config, nil)
}

//func TestInvokePostAndGetCallsForAllObjects(t *testing.T) {
//	// given
//	config := testdoubles.LoadTestConfig(t)
//
//	template := environment.Template{Content:templateHeader+namespaceObject+roleBindingRestrictionObject}
//	objects, err := template.Process(environment.CollectVars("aslak-test", config))
//	require.NoError(t, err)
//
//	gock.New("http://starter.com").
//		Post("/api/v1/namespaces").
//		Reply(200)
//	gock.New("http://starter.com").
//		Get("/api/v1/namespaces/aslak-test").
//		Reply(200)
//	gock.New("http://starter.com").
//		Post("/oapi/v1/namespaces/aslak-test/rolebindingrestrictions").
//		Reply(200)
//	gock.New("http://starter.com").
//		Get("/oapi/v1/namespaces/aslak-test/rolebindingrestrictions/dsaas-user-access").
//		Reply(200)
//
//	// when
//	err = createOpenshiftClient(config).ApplyAll(objects).WithPostMethod()
//
//	// then
//	require.NoError(t, err)
//}
//
//func TestDeleteIfThereIsConflict(t *testing.T) {
//	// given
//	config := testdoubles.LoadTestConfig(t)
//	template := environment.Template{Content:templateHeader+roleBindingRestrictionObject}
//	objects, err := template.Process(environment.CollectVars("aslak-test", config))
//
//	require.NoError(t, err)
//
//	gock.New("http://starter.com").
//		Post("/oapi/v1/namespaces/aslak-test/rolebindingrestrictions").
//		Reply(409)
//	gock.New("http://starter.com").
//		Delete("/oapi/v1/namespaces/aslak-test/rolebindingrestrictions/dsaas-user-access").
//		Reply(200)
//	gock.New("http://starter.com").
//		Get("/oapi/v1/namespaces/aslak-test/rolebindingrestrictions/dsaas-user-access").
//		Reply(404)
//	gock.New("http://starter.com").
//		Post("/oapi/v1/namespaces/aslak-test/rolebindingrestrictions").
//		Reply(200)
//	gock.New("http://starter.com").
//		Get("/oapi/v1/namespaces/aslak-test/rolebindingrestrictions/dsaas-user-access").
//		Reply(200)
//
//	// when
//	err = createOpenshiftClient(config).ApplyAll(objects).WithPostMethod()
//
//	// then
//	require.NoError(t, err)
//}
//
//func TestDeleteAndGet(t *testing.T) {
//	// given
//	config := testdoubles.LoadTestConfig(t)
//	template := environment.Template{Content:templateHeader+roleBindingRestrictionObject}
//	objects, err := template.Process(environment.CollectVars("aslak-test", config))
//
//	require.NoError(t, err)
//
//	gock.New("http://starter.com").
//		Delete("/oapi/v1/namespaces/aslak-test/rolebindingrestrictions/dsaas-user-access").
//		Reply(200)
//	gock.New("http://starter.com").
//		Get("/oapi/v1/namespaces/aslak-test/rolebindingrestrictions/dsaas-user-access").
//		Reply(404)
//
//	// when
//	err = createOpenshiftClient(config).ApplyAll(objects).WithDeleteMethod()
//
//	// then
//	require.NoError(t, err)
//}