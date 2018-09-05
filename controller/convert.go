package controller

import (
	"github.com/fabric8-services/fabric8-tenant/dbsupport"
	"github.com/fabric8-services/fabric8-tenant/app"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/sirupsen/logrus"
	"context"
	"github.com/fabric8-services/fabric8-tenant/utils"
	"github.com/satori/go.uuid"
)

type NamespaceFilter func(namespace dbsupport.Namespace) bool

func convertTenant(tenant *dbsupport.Tenant, service cluster.Service, ctx context.Context, filterNs NamespaceFilter) *app.TenantSingle {
	namespaces := make([]*app.NamespaceAttributes, 0)

	for _, ns := range tenant.Namespaces {
		if !filterNs(ns) {
			continue
		}

		cluster, err := service.GetCluster(ctx, ns.ClusterURL)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":         err,
				"cluster_url": ns.ClusterURL,
			}).Error("unable to resolve cluster")
			//todo handle error
		}
		namespaces = append(namespaces, &app.NamespaceAttributes{
			CreatedAt:         &ns.CreatedAt,
			UpdatedAt:         &ns.UpdatedAt,
			ClusterURL:        &ns.ClusterURL,
			ClusterAppDomain:  &cluster.AppDNS,
			ClusterConsoleURL: &cluster.ConsoleURL,
			ClusterMetricsURL: &cluster.MetricsURL,
			ClusterLoggingURL: &cluster.LoggingURL,
			Name:              &ns.Name,
			Type:              &ns.Type,
			Version:           &ns.TemplateVersion,
			State:             utils.String(string(ns.State)),
			ClusterCapacityExhausted: &cluster.CapacityExhausted,
		})
	}

	return &app.TenantSingle{
		Data: &app.Tenants{
			ID:   &tenant.ID,
			Type: "tenants",
			Attributes: &app.TenantAttributes{
				CreatedAt:  &tenant.CreatedAt,
				Namespaces: namespaces,
			},
		},
	}
}

func filterByNsAndSpace(nsType *string, space *uuid.UUID) NamespaceFilter {
	return func(namespace dbsupport.Namespace) bool {
		return filterByNs(nsType)(namespace) && filterBySpace(space)(namespace)
	}
}

func filterByNs(nsType *string) NamespaceFilter {
	return func(namespace dbsupport.Namespace) bool {
		return utils.IsEmpty(nsType) || namespace.Type == *nsType
	}
}

func filterBySpace(space *uuid.UUID) NamespaceFilter {
	return func(namespace dbsupport.Namespace) bool {
		return space == nil || namespace.SpaceID == utils.UuidValue(space)
	}
}

func asMap(slice []string) map[string]string {
	var asMap map[string]string
	for _, value := range slice {
		asMap[value] = value
	}
	return asMap
}
