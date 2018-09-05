package dbsupport

import (
	"time"
	"github.com/satori/go.uuid"
	"strconv"
	"github.com/jinzhu/gorm"
	"github.com/goadesign/goa"
	"github.com/fabric8-services/fabric8-tenant/cluster"
	"github.com/fabric8-services/fabric8-tenant/environment"
	"github.com/fabric8-services/fabric8-tenant/utils"
	log "github.com/sirupsen/logrus"
	errs "github.com/pkg/errors"
	"github.com/fabric8-services/fabric8-common/errors"
	"fmt"
)

type NamespaceState string

const (
	Provisioning NamespaceState = "provisioning"
	Updating     NamespaceState = "updating"
	Ready        NamespaceState = "ready"
	Failed       NamespaceState = "failed"
)

// Namespace represent a single namespace owned by an Tenant
type Namespace struct {
	Lifecycle
	ID uuid.UUID `sql:"type:uuid default uuid_generate_v4()"`

	IdentityID uuid.UUID `gorm:"primary_key;column:identity_id" sql:"type:uuid"`
	SpaceID    uuid.UUID `gorm:"primary_key;column:space_id" sql:"type:uuid"`
	Type       string    `gorm:"primary_key;column:type"`

	NamespaceName string
	Name          string

	ClusterURL      string
	ClusterUsername string

	TemplateName    string
	TemplateVersion string

	ExpiresAt *time.Time
	State     NamespaceState
}

// TableName overrides the table name settings in Gorm to force a specific table name
// in the database.
func (n *Namespace) TableName() string {
	return "namespaces"
}

// GetETagData returns the field values to use to generate the ETag
func (n *Namespace) GetETagData() []interface{} {
	// using the 'ID' and 'UpdatedAt' (converted to number of seconds since epoch) fields
	return []interface{}{n.ID, strconv.FormatInt(n.UpdatedAt.Unix(), 10)}
}

func (n *Namespace) UpdateData(env *environment.EnvData, cluster *cluster.Cluster, state NamespaceState) {
	n.NamespaceName = env.NamespaceName
	n.Name = env.Name
	n.TemplateName = env.Template.Filename
	n.TemplateVersion = env.Template.Version
	n.ClusterURL = cluster.APIURL
	n.ClusterUsername = cluster.User
	n.ExpiresAt = env.ExpiresAt
	n.State = state
}

// GetLastModified returns the last modification time
func (m Namespace) GetLastModified() time.Time {
	return m.UpdatedAt
}

type NamespaceRepository interface {
	//Exists(tenantID string) bool
	//Load(tenantID string) (*Tenant, error)
	Load(space *uuid.UUID, nsType string) (*Namespace, error)
	Create(state *Namespace) error   // CreateTenant will return err on duplicate insert
	Save(namespace *Namespace) error // SaveTenant will update on dupliate 'insert'
	Delete(namespace *Namespace) error
	NewNamespace(space *uuid.UUID, nsType string) *Namespace
}

func NewNamespaceRepository(db *gorm.DB, tenantID uuid.UUID) NamespaceRepository {
	return &GormNamespaceRepository{db: db, tenantID: tenantID}
}

type GormNamespaceRepository struct {
	db       *gorm.DB
	tenantID uuid.UUID
}

func (r *GormNamespaceRepository) NewNamespace(space *uuid.UUID, nsType string) *Namespace {
	return &Namespace{
		IdentityID: r.tenantID,
		SpaceID:    utils.UuidValue(space),
		Type:       nsType,
	}
}

func (r *GormNamespaceRepository) Save(namespace *Namespace) error {
	defer goa.MeasureSince([]string{"goa", "db", "namespace", "save"}, time.Now())
	return r.db.Save(namespace).Error
}

func (r *GormNamespaceRepository) Create(namespace *Namespace) error {
	defer goa.MeasureSince([]string{"goa", "db", "namespace", "create"}, time.Now())
	return r.db.Create(namespace).Error
}

// Delete removes a single record.
func (r *GormNamespaceRepository) Delete(namespace *Namespace) error {
	defer goa.MeasureSince([]string{"goa", "db", "namespace", "delete"}, time.Now())

	err := r.db.Delete(namespace).Error

	if err != nil {
		log.WithFields(log.Fields{
			"tenant_id": namespace.IdentityID,
			"space":     namespace.SpaceID,
			"type":      namespace.Type,
			"err":       err,
		}).Error("unable to delete the tenant")
		return errs.WithStack(err)
	}

	log.WithFields(log.Fields{
		"tenant_id": namespace.IdentityID,
		"space":     namespace.SpaceID,
		"type":      namespace.Type,
	}).Debug("Tenant deleted!")

	return nil
}


func (r *GormNamespaceRepository) Load(space *uuid.UUID, nsType string) (*Namespace, error) {
	var ns Namespace
	err := r.db.Table(ns.TableName()).
		Where("identity_id = ? AND space_id = ? AND type = ?", r.tenantID, space, nsType).Find(&ns).Error

	if err == gorm.ErrRecordNotFound {
		entityId := fmt.Sprintf("identity_id=%s & space_id=%s & type=%s", r.tenantID, space, nsType)
		return nil, errors.NewNotFoundError("namespace", entityId)

	} else if err != nil {
		return nil, errs.Wrapf(err, "unable to lookup namespace")
	}
	return &ns, nil
}