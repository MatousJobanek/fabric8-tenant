package controller

import (
	"context"
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/fabric8-services/fabric8-tenant/tenant"
	"github.com/fabric8-services/fabric8-tenant/utils"
	"github.com/sirupsen/logrus"
)

type NamespaceFilter func(namespace tenant.Namespace) bool
type ClusterResolver func(ctx context.Context, target string) (cluster.Cluster, error)

func convertTenant(ctx context.Context, tenant *tenant.Tenant, namespaces []*tenant.Namespace, resolveCluster ClusterResolver) *app.Tenant {
	nsAttributes := make([]*app.NamespaceAttributes, 0)

	for _, ns := range namespaces {

		nsCluster, err := resolveCluster(ctx, ns.MasterURL)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":         err,
				"cluster_url": ns.MasterURL,
			}).Error("unable to resolve nsCluster")
			nsCluster = cluster.Cluster{}
		}
		nsAttributes = append(nsAttributes, &app.NamespaceAttributes{
			CreatedAt:         &ns.CreatedAt,
			UpdatedAt:         &ns.UpdatedAt,
			ClusterURL:        &ns.MasterURL,
			ClusterAppDomain:  &nsCluster.AppDNS,
			ClusterConsoleURL: &nsCluster.ConsoleURL,
			ClusterMetricsURL: &nsCluster.MetricsURL,
			ClusterLoggingURL: &nsCluster.LoggingURL,
			Name:              &ns.Name,
			Type:              utils.String(ns.Type.String()),
			Version:           &ns.Version,
			State:             utils.String(ns.State.String()),
			ClusterCapacityExhausted: &nsCluster.CapacityExhausted,
		})
	}

	return &app.Tenant{
		ID:   &tenant.ID,
		Type: "tenants",
		Attributes: &app.TenantAttributes{
			CreatedAt:  &tenant.CreatedAt,
			Email:      &tenant.Email,
			Profile:    &tenant.Profile,
			Namespaces: nsAttributes,
		},
	}
}

//func filterByNsAndSpace(nsType *string, space *uuid.UUID) NamespaceFilter {
//	return func(namespace tenant.Namespace) bool {
//		return filterByNs(nsType)(namespace) && filterBySpace(space)(namespace)
//	}
//}
//
//func filterByNs(nsType *string) NamespaceFilter {
//	return func(namespace tenant.Namespace) bool {
//		return utils.IsEmpty(nsType) || namespace.Type == *nsType
//	}
//}
//
//func filterBySpace(space *uuid.UUID) NamespaceFilter {
//	return func(namespace tenant.Namespace) bool {
//		return space == nil || namespace.SpaceID == utils.UuidValue(space)
//	}
//}

func asMap(slice []string) map[string]string {
	var asMap map[string]string
	for _, value := range slice {
		asMap[value] = value
	}
	return asMap
}
