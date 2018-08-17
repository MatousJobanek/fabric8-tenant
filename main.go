//go:generate goagen bootstrap -d github.com/fabric8-services/fabric8-tenant/design

package main

import (
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"github.com/fabric8-services/fabric8-tenant/controller"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/sirupsen/logrus"
	"github.com/fabric8-services/fabric8-tenant/log"
	"github.com/fabric8-services/fabric8-tenant/configuration"
	witmiddleware "github.com/fabric8-services/fabric8-wit/goamiddleware"
	"flag"
	"github.com/spf13/viper"
	"github.com/fabric8-services/fabric8-tenant/jsonapi"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/fabric8-services/fabric8-tenant/auth"
)

func main() {

	viper.GetStringMapString("TEST")

	//var migrateDB bool
	//flag.BoolVar(&migrateDB, "migrateDatabase", false, "Migrates the database to the newest version and exits.")
	flag.Parse()

	// Initialized configuration
	config, err := configuration.GetData()
	if err != nil {
		logrus.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed to setup the configuration")
	}

	log := configureLogger(config)

	// Create service
	service := goa.New("tenant")

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(true))
	service.Use(jsonapi.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	authService, err := auth.NewAuthService(config, log)
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed to fetch service account token")
	}

	publicKeys, err := authService.GetPublicKeys()
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"err":    err,
			"target": config.GetAuthURL(),
		}, "failed to fetch public keys from token service")
	}

	service.Use(witmiddleware.TokenContext([]string{"secret"}, nil, app.NewJWTSecurity()))
	app.UseJWTMiddleware(service, goajwt.New([]string{"secret"}, nil, app.NewJWTSecurity()))

	service.Use(witmiddleware.TokenContext(publicKeys, nil, app.NewJWTSecurity()))
	//service.Use(log.LogRequest(config.IsDeveloperModeEnabled()))
	app.UseJWTMiddleware(service, goajwt.New(publicKeys, nil, app.NewJWTSecurity()))

	// Mount "status" controller
	c := controller.NewStatusController(service)
	app.MountStatusController(service, c)

	//cluster := controller.Cluster{
	//	APIURL:"https://192.168.42.241:8443",
	//	Token:"3PhkSX3hqmHyk1XXuFjL5-xzvV9iG1-BiPAvij7jxwg",
	//}

	clusterService, err := cluster.NewClusterService(config.GetClustersRefreshDelay(), authService)
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed to initialize the cluster.Service component")
	}
	defer clusterService.Stop()

	// Mount "tenant" controller
	tenant := controller.NewTenantController(service, clusterService, authService, log, config)
	app.MountTenantController(service, tenant)
	// Mount "tenants" controller
	tenants := controller.NewTenantsController(service)
	app.MountTenantsController(service, tenants)

	// Start service
	if err := service.ListenAndServe(":8080"); err != nil {
		service.LogError("startup", "err", err)
	}
}

func configureLogger(config *configuration.Data) *logrus.Entry {
	logger := log.ConfigureLogrus()

	//rawSecret, err := ioutil.ReadFile(secretFilename)
	//if err != nil {
	//	logger.WithError(err).Errorf("unable to load sentry dsn from %q. No sentry integration enabled", *sentryDsnSecretFile)
	//} else {
	//	version, found := os.LookupEnv("VERSION")
	//	if !found {
	//		version = "UNKNOWN"
	//	}
	//	log.AddSentryHook(logger, log.NewSentryConfiguration(string(bytes.TrimSpace(rawSecret)), map[string]string{
	//		"plugin":      pluginName,
	//		"environment": *environment,
	//		"version":     version,
	//	}, *sentryTimeout))
	//}

	return logger
}
