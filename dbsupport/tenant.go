package dbsupport

import (
	"github.com/satori/go.uuid"
	"github.com/jinzhu/gorm"
	"strconv"
	"time"
	"fmt"
	"github.com/goadesign/goa"
	errs "github.com/pkg/errors"
	"github.com/fabric8-services/fabric8-common/errors"
	log "github.com/sirupsen/logrus"
)

// Tenant is the owning OpenShift account
type Tenant struct {
	Lifecycle

	ID         uuid.UUID
	Email      string `sql:"unique_index"`
	OSUsername string

	Namespaces []Namespace `gorm:"foreignkey:IdentityID;association_foreignkey:ID"`
}

// TableName overrides the table name settings in Gorm to force a specific table name
// in the database.
func (m Tenant) TableName() string {
	return "tenants"
}

// GetETagData returns the field values to use to generate the ETag
func (m Tenant) GetETagData() []interface{} {
	// using the 'ID' and 'UpdatedAt' (converted to number of seconds since epoch) fields
	return []interface{}{m.ID, strconv.FormatInt(m.UpdatedAt.Unix(), 10)}
}

// GetLastModified returns the last modification time
func (m Tenant) GetLastModified() time.Time {
	return m.UpdatedAt
}

func (t *Tenant) GetNamespaceByType(space uuid.UUID) map[string]Namespace {
	var nsTypes map[string]Namespace
	for _, namespace := range t.Namespaces {
		if namespace.SpaceID == space {
			nsTypes[namespace.Type] = namespace
		}
	}
	return nsTypes
}

func (t *Tenant) GetNamespacesBySpace() map[uuid.UUID][]Namespace {
	var namespacesBySpace map[uuid.UUID][]Namespace
	for _, namespace := range t.Namespaces {
		namespacesBySpace[namespace.SpaceID] = append(namespacesBySpace[namespace.SpaceID], namespace)
	}
	return namespacesBySpace
}

func (t *Tenant) GetNamespace(nsType string, space uuid.UUID) (Namespace, bool) {
	for _, namespace := range t.Namespaces {
		if namespace.Type == nsType && namespace.SpaceID == space {
			return namespace, true
		}
	}
	return Namespace{}, false
}

type TenantRepository interface {
	Exists(tenantID string) bool
	Load(tenantID string) (*Tenant, error)
	Lookup(masterURL, namespace string) (*Tenant, error)
	Create(tenant *Tenant) error // CreateTenant will return err on duplicate insert
	Save(tenant *Tenant) error   // SaveTenant will update on dupliate 'insert'
	Delete(tenantID uuid.UUID) error
	NewNamespaceRepository(tenantID uuid.UUID) NamespaceRepository
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &GormTenantRepository{db: db}
}

type GormTenantRepository struct {
	db *gorm.DB
}

func (r GormTenantRepository) NewNamespaceRepository(tenantID uuid.UUID) NamespaceRepository {
	return NewNamespaceRepository(r.db, tenantID)
}

func (r GormTenantRepository) Exists(tenantID string) bool {
	defer goa.MeasureSince([]string{"goa", "db", "tenant", "exists"}, time.Now())
	var t Tenant
	err := r.db.Table(t.TableName()).Where("id = ?", tenantID).Find(&t).Error
	if err != nil {
		return false
	}
	return true
}

func (r GormTenantRepository) Load(tenantID string) (*Tenant, error) {
	defer goa.MeasureSince([]string{"goa", "db", "tenant", "load"}, time.Now())
	var t Tenant
	err := r.db.Table(t.TableName()).Where("id = ?", tenantID).Find(&t).Error
	if err == gorm.ErrRecordNotFound {
		// no match
		return nil, errors.NewNotFoundError("tenant", tenantID)
	} else if err != nil {
		return nil, errs.Wrapf(err, "unable to lookup tenant by id")
	}
	return &t, nil
}

func (r GormTenantRepository) Lookup(masterURL, namespace string) (*Tenant, error) {
	defer goa.MeasureSince([]string{"goa", "db", "tenant", "lookup"}, time.Now())
	// select t.id from tenant t, namespaces n where t.id = n.tenant_id and n.master_url = ? and n.name = ?
	var result Tenant
	query := fmt.Sprintf("select t.* from %[1]s t, %[2]s n where t.id = n.tenant_id and n.master_url = ? and n.name = ?", (&result).TableName(), (&Namespace{}).TableName())
	err := r.db.Raw(query, masterURL, namespace).Scan(&result).Error
	if err == gorm.ErrRecordNotFound {
		// no match
		return nil, errors.NewNotFoundError("tenant", "")
	} else if err != nil {
		return nil, errs.Wrapf(err, "unable to lookup tenant by namespace")
	}
	return &result, nil
}

func (r GormTenantRepository) Save(tenant *Tenant) error {
	defer goa.MeasureSince([]string{"goa", "db", "tenant", "save"}, time.Now())
	return r.db.Save(tenant).Error
}

func (r GormTenantRepository) Create(tenant *Tenant) error {
	defer goa.MeasureSince([]string{"goa", "db", "tenant", "save"}, time.Now())
	return r.db.Create(tenant).Error
}

// Delete removes a single record.
func (r *GormTenantRepository) Delete(id uuid.UUID) error {
	defer goa.MeasureSince([]string{"goa", "db", "user", "delete"}, time.Now())

	obj := Tenant{ID: id}

	err := r.db.Delete(&obj).Error

	if err != nil {
		log.WithFields(log.Fields{
			"tenant_id": id,
			"err":       err,
		}).Error("unable to delete the tenant")
		return errs.WithStack(err)
	}

	log.WithFields(log.Fields{
		"tenant_id": id,
	}).Debug("Tenant deleted!")

	return nil
}

//func (s DBService) SaveNamespace(namespace *Namespace) error {
//	if namespace.ID == uuid.Nil {
//		namespace.ID = uuid.NewV4()
//	}
//	return s.db.Save(namespace).Error
//}
//
//func (s DBService) GetNamespaces(tenantID uuid.UUID) ([]*Namespace, error) {
//	var t []*Namespace
//	err := s.db.Table(Namespace{}.TableName()).Where("tenant_id = ?", tenantID).Find(&t).Error
//	if err != nil {
//		return nil, err
//	}
//	return t, nil
//}
//
//func (s DBService) DeleteAll(tenantID uuid.UUID) error {
//	err := s.deleteNamespaces(tenantID)
//	err = s.deleteTenant(tenantID)
//	return err
//}
//
//func (s DBService) deleteNamespaces(tenantID uuid.UUID) error {
//	if tenantID == uuid.Nil {
//		return nil
//	}
//	return s.db.Unscoped().Delete(&Namespace{}, "tenant_id = ?", tenantID).Error
//}
//
//func (s DBService) deleteTenant(tenantID uuid.UUID) error {
//	if tenantID == uuid.Nil {
//		return nil
//	}
//	return s.db.Unscoped().Delete(&Tenant{ID: tenantID}).Error
//}
