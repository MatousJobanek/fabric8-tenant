package update_test

import (
	"context"
	"fmt"
	goatest "github.com/fabric8-services/fabric8-tenant/app/test"
	"github.com/fabric8-services/fabric8-tenant/controller"
	"github.com/fabric8-services/fabric8-tenant/tenant"
	"github.com/fabric8-services/fabric8-tenant/test"
	"github.com/fabric8-services/fabric8-tenant/test/doubles"
	"github.com/fabric8-services/fabric8-tenant/test/gormsupport"
	"github.com/fabric8-services/fabric8-tenant/test/minishift"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"github.com/fabric8-services/fabric8-tenant/update"
	"sync"
	"os"
)

type AutomatedUpdateMinishiftTestSuite struct {
	minishift.TestSuite
}

func TestAutomatedUpdateWithMinishift(t *testing.T) {
	os.Setenv("F8_RESOURCE_DATABASE", "1")
	os.Setenv("F8_POSTGRES_DATABASE", "postgres")
	os.Setenv("F8_MINISHIFT", "1")
	os.Setenv("F8_MINISHIFT_USER_NAME", "developer")
	os.Setenv("F8_MINISHIFT_USER_TOKEN", "bPxToO17WBqMJKetH4Ulk4tdEslME9FxPUWmv-8qxKU")
	os.Setenv("F8_MINISHIFT_ADMIN_NAME", "admin")
	os.Setenv("F8_MINISHIFT_ADMIN_TOKEN", "UtF6bO1eGqXnLYZdWr4LJrsziWU03RnkNRIWbqheQdk")
	os.Setenv("F8_MINISHIFT_URL", "https://192.168.42.95:8443")
	toReset := test.SetEnvironments(test.Env("F8_AUTOMATED_UPDATE_RETRY_SLEEP", (5 * time.Second).String()))
	defer toReset()

	suite.Run(t, &AutomatedUpdateMinishiftTestSuite{
		TestSuite: minishift.TestSuite{DBTestSuite: gormsupport.NewDBTestSuite("../config.yaml")}})
}

func (s *AutomatedUpdateMinishiftTestSuite) TestSetupUpdateCleanAndDeleteTenantNamespaces() {
	// given
	testdoubles.SetTemplateSameVersion("1abcd")
	svc := goa.New("Tenants-service")
	var tenantIDs []uuid.UUID

	for i := 0; i < 2; i++ {
		id := uuid.NewV4()
		tenantIDs = append(tenantIDs, id)
		ctrl := controller.NewTenantController(svc, tenant.NewDBService(s.DB), s.GetClusterService(), s.GetAuthService(id), s.GetConfig())
		goatest.SetupTenantAccepted(s.T(), createUserContext(s.T(), id.String()), svc, ctrl)

	}

	for _, tenantID := range tenantIDs {
		repo := tenant.NewDBService(s.DB)
		iteration := 0
		for {
			namespaces, err := repo.GetNamespaces(tenantID)
			if err != nil {
				assert.NoError(s.T(), err)
				break
			}
			if len(namespaces) == 5 {
				break
			}
			if iteration == 20 {
				assert.Fail(s.T(), fmt.Sprintf("not all namespaces created. created only: %+v", namespaces))
				break
			}
			iteration++
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Println("created ", tenantID)
	}
	defer s.clean(tenantIDs)

	tx(s.T(), s.DB, func(repo update.Repository) error {
		if err := repo.UpdateStatus(update.Finished); err != nil {
			return err
		}
		return updateVersionsTo(repo, "1abcd")
	})
	before := time.Now()

	// when
	testdoubles.SetTemplateSameVersion("2abcd")

	var waitToContinueGoroutine sync.WaitGroup
	waitToContinueGoroutine.Add(1)
	var waitToFinish sync.WaitGroup
	updateExec := DummyUpdateExecutor{shouldCallOriginalUpdater: true, numberOfCalls: Uint64(0)}
	for i := 0; i < 10; i++ {
		waitToFinish.Add(1)
		go func(toWait *sync.WaitGroup, toMarkAsDone *sync.WaitGroup, updateExecutor controller.UpdateExecutor) {
			defer toMarkAsDone.Done()

			toWait.Wait()
			update.NewTenantsUpdater(s.DB, s.Config, s.AuthService, s.ClusterService, updateExecutor).UpdateAllTenants()
		}(&waitToContinueGoroutine, &waitToFinish, &updateExec)
	}
	fmt.Println("making done")
	waitToContinueGoroutine.Done()
	fmt.Println("waiting for to finish")
	waitToFinish.Wait()
	fmt.Println("finished")
	// then
	assert.Equal(s.T(), 5*len(tenantIDs), int(*updateExec.numberOfCalls))
	fmt.Println("going to verify")
	s.verifyAreUpdated(tenantIDs, before)
}

func (s *AutomatedUpdateMinishiftTestSuite) verifyAreUpdated(tenantIDs []uuid.UUID, wasBefore time.Time) {
	var wg sync.WaitGroup
	for _, tenantID := range tenantIDs {
		wg.Add(1)
		go func(toMarkAsDone *sync.WaitGroup, t *testing.T) {
			defer toMarkAsDone.Done()
			repo := tenant.NewDBService(s.DB)
			tnnt, err := repo.GetTenant(tenantID)
			assert.NoError(t, err)
			namespaces, err := repo.GetNamespaces(tenantID)
			assert.NoError(t, err)
			assert.Len(t, namespaces, 5)
			for _, ns := range namespaces {
				assert.True(t, wasBefore.Before(ns.UpdatedAt))
				//assert.Equal(t, "2abcd", ns.Version)
				assert.Equal(t, "ready", ns.State)
			}
			fmt.Println("namespaces done", tenantID)
			mappedObjects, masterOpts := s.GetMappedTemplateObjects(tnnt.NsBaseName)
			minishift.VerifyObjectsPresence(t, mappedObjects, masterOpts, "2abcd")
			fmt.Println("objects done", tenantID)
		}(&wg, s.T())
	}
	wg.Wait()
}

func (s *AutomatedUpdateMinishiftTestSuite) clean(toCleanup []uuid.UUID) {
	fmt.Println("going to cleanup")
	svc := goa.New("Tenants-service")
	var wg sync.WaitGroup
	for _, tenantID := range toCleanup {
		wg.Add(1)
		go func(toMarkAsDone *sync.WaitGroup) {
			ctrl := controller.NewTenantController(svc, tenant.NewDBService(s.DB), s.GetClusterService(), s.GetAuthService(tenantID), s.GetConfig())
			goatest.CleanTenantNoContent(s.T(), createUserContext(s.T(), tenantID.String()), svc, ctrl, true)
			fmt.Println("cleaned", tenantID)
		}(&wg)
	}
	wg.Wait()
}

func createUserContext(t *testing.T, sub string) context.Context {
	userToken, err := test.NewToken(
		map[string]interface{}{
			"sub":                sub,
			"preferred_username": "developer",
			"email":              "developer@redhat.com",
		},
		"../test/private_key.pem",
	)
	require.NoError(t, err)

	return goajwt.WithJWT(context.Background(), userToken)
}
