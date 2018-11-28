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
)

type AutomatedUpdateMinishiftTestSuite struct {
	minishift.TestSuite
}

func TestAutomatedUpdateWithMinishift(t *testing.T) {
	//os.Setenv("F8_RESOURCE_DATABASE", "1")
	//os.Setenv("F8_POSTGRES_DATABASE", "postgres")
	//os.Setenv("F8_MINISHIFT", "1")
	//os.Setenv("F8_MINISHIFT_USER_NAME", "developer18.11.28.25")
	//os.Setenv("F8_MINISHIFT_USER_TOKEN", "fgAM61Rhv6KuVD7QOSA-aKcfmJmSF5eFqtX5_vW0H38")
	//os.Setenv("F8_MINISHIFT_ADMIN_NAME", "admin")
	//os.Setenv("F8_MINISHIFT_ADMIN_TOKEN", "EV9D9AYK4ppmoDxw4VMmZ9XgBHkhoz7Fas48sexf3kA")
	//os.Setenv("F8_MINISHIFT_URL", "https://192.168.42.231:8443")
	toReset := test.SetEnvironments(test.Env("F8_AUTOMATED_UPDATE_RETRY_SLEEP", (5 * time.Second).String()))
	defer toReset()

	suite.Run(t, &AutomatedUpdateMinishiftTestSuite{
		TestSuite: minishift.TestSuite{DBTestSuite: gormsupport.NewDBTestSuite("../config.yaml")}})
}

func (s *AutomatedUpdateMinishiftTestSuite) TestAutomaticUpdateOfTenantNamespaces() {
	// given
	testdoubles.SetTemplateSameVersion("1abcd")
	svc := goa.New("Tenants-service")
	var tenantIDs []uuid.UUID
	clusterService := s.GetClusterService()

	for i := 0; i < 4; i++ {
		id := uuid.NewV4()
		tenantIDs = append(tenantIDs, id)
		ctrl := controller.NewTenantController(svc, tenant.NewDBService(s.DB), clusterService, s.GetAuthService(id), s.GetConfig())
		goatest.SetupTenantAccepted(s.T(), createUserContext(s.T(), id.String()), svc, ctrl)
	}

	for _, tenantID := range tenantIDs {
		repo := tenant.NewDBService(s.DB)
		err := test.WaitWithTimeout(30 * time.Second).Until(func() error {
			namespaces, err := repo.GetNamespaces(tenantID)
			if err != nil {
				return err
			}
			if len(namespaces) != 5 {
				return fmt.Errorf("not all namespaces created. created only: %+v", namespaces)
			}
			for _, ns := range namespaces {
				fmt.Println(ns.Name)
			}
			return nil
		})
		require.NoError(s.T(), err)
		tnnt, err := repo.GetTenant(tenantID)
		require.NoError(s.T(), err)
		mappedObjects, masterOpts := s.GetMappedTemplateObjects(tnnt.NsBaseName)
		minishift.VerifyObjectsPresence(s.T(), mappedObjects, masterOpts, "1abcd", true)
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

	var goroutineCanContinue sync.WaitGroup
	goroutineCanContinue.Add(1)
	var goroutineFinished sync.WaitGroup
	updateExec := DummyUpdateExecutor{shouldCallOriginalUpdater: true, numberOfCalls: Uint64(0)}
	for i := 0; i < 10; i++ {
		goroutineFinished.Add(1)
		go func(updateExecutor controller.UpdateExecutor) {
			defer goroutineFinished.Done()

			goroutineCanContinue.Wait()
			update.NewTenantsUpdater(s.DB, s.Config, clusterService, updateExecutor).UpdateAllTenants()
		}(&updateExec)
	}
	fmt.Println("making done")
	goroutineCanContinue.Done()
	fmt.Println("waiting for to finish")
	goroutineFinished.Wait()
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
		go func(t *testing.T, tenantID uuid.UUID) {
			defer wg.Done()
			repo := tenant.NewDBService(s.DB)
			tnnt, err := repo.GetTenant(tenantID)
			assert.NoError(t, err)
			namespaces, err := repo.GetNamespaces(tenantID)
			assert.NoError(t, err)
			assert.Len(t, namespaces, 5)
			for _, ns := range namespaces {
				assert.True(t, wasBefore.Before(ns.UpdatedAt))
				assert.Contains(t, ns.Version, "2abcd")
				assert.NotContains(t, ns.Version, "1abcd")
				assert.Equal(t, "ready", ns.State)
			}
			fmt.Println("namespaces done", tenantID)
			mappedObjects, masterOpts := s.GetMappedTemplateObjects(tnnt.NsBaseName)
			minishift.VerifyObjectsPresence(t, mappedObjects, masterOpts, "2abcd", false)
			fmt.Println("objects done", tenantID)
		}(s.T(), tenantID)
	}
	wg.Wait()
}

func (s *AutomatedUpdateMinishiftTestSuite) clean(toCleanup []uuid.UUID) {
	fmt.Println("going to cleanup")
	svc := goa.New("Tenants-service")
	var wg sync.WaitGroup
	for _, tenantID := range toCleanup {
		wg.Add(1)
		go func(tenantID uuid.UUID) {
			defer wg.Done()
			ctrl := controller.NewTenantController(svc, tenant.NewDBService(s.DB), s.GetClusterService(), s.GetAuthService(tenantID), s.GetConfig())
			goatest.CleanTenantNoContent(s.T(), createUserContext(s.T(), tenantID.String()), svc, ctrl, true)
			fmt.Println("cleaned", tenantID)
		}(tenantID)
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
