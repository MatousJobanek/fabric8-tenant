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

	service.Use(witmiddleware.TokenContext([]string{"secret"}, nil, app.NewJWTSecurity()))
	app.UseJWTMiddleware(service, goajwt.New([]string{"secret"}, nil, app.NewJWTSecurity()))

	// Mount "status" controller
	c := controller.NewStatusController(service)
	app.MountStatusController(service, c)

	cluster := controller.Cluster{
		APIURL:"https://192.168.42.241:8443",
		Token:"3PhkSX3hqmHyk1XXuFjL5-xzvV9iG1-BiPAvij7jxwg",
	}

	// Mount "tenant" controller
	tenant := controller.NewTenantController(service, cluster, log, config)
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
