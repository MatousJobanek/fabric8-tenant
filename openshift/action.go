package openshift

import (
	"net/http"
	"github.com/fabric8-services/fabric8-tenant/dbsupport"
	log "github.com/sirupsen/logrus"
	"github.com/fabric8-services/fabric8-tenant/environment"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/satori/go.uuid"
)

type NamespaceAction interface {
	methodName() string
	getNamespaceEntity(space *uuid.UUID, nsType string, nsRepo dbsupport.NamespaceRepository) (*dbsupport.Namespace, error)
	updateTable(env *environment.EnvData, cluster *cluster.Cluster, namespace *dbsupport.Namespace, nsRepo dbsupport.NamespaceRepository)
}

type create struct {
}

func (c *create) methodName() string {
	return http.MethodPost
}

func (c *create) getNamespaceEntity(space *uuid.UUID, nsType string, nsRepo dbsupport.NamespaceRepository) (*dbsupport.Namespace, error) {
	namespace := nsRepo.NewNamespace(space, nsType)
	namespace.State = dbsupport.Provisioning
	err := nsRepo.Create(namespace)
	return namespace, err
}

func (c *create) updateTable(env *environment.EnvData, cluster *cluster.Cluster, namespace *dbsupport.Namespace, nsRepo dbsupport.NamespaceRepository) {
	namespace.UpdateData(env, cluster, dbsupport.Ready)
	err := nsRepo.Save(namespace)
	if err != nil {
		log.Error(err)
	}
}

type delete struct {
}

func (d *delete) methodName() string {
	return http.MethodDelete
}

func (d *delete) getNamespaceEntity(space *uuid.UUID, nsType string, nsRepo dbsupport.NamespaceRepository) (*dbsupport.Namespace, error) {
	return nsRepo.NewNamespace(space, nsType), nil
}

func (d *delete) updateTable(env *environment.EnvData, cluster *cluster.Cluster, namespace *dbsupport.Namespace, nsRepo dbsupport.NamespaceRepository) {
	err := nsRepo.Delete(namespace)
	if err != nil {
		log.Error(err)
	}
}

type update struct {
}

func (u *update) methodName() string {
	return http.MethodPatch
}

func (u *update) getNamespaceEntity(space *uuid.UUID, nsType string, nsRepo dbsupport.NamespaceRepository) (*dbsupport.Namespace, error) {
	return nsRepo.Load(space, nsType)
}

func (u *update) updateTable(env *environment.EnvData, cluster *cluster.Cluster, namespace *dbsupport.Namespace, nsRepo dbsupport.NamespaceRepository) {
	namespace.UpdateData(env, cluster, dbsupport.Ready)
	err := nsRepo.Save(namespace)
	if err != nil {
		log.Error(err)
	}
}