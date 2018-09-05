package dbsupport_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/fabric8-services/fabric8-tenant/dbsupport"
	"github.com/fabric8-services/fabric8-tenant/test/resource"
	"github.com/fabric8-services/fabric8-common/convert"
)

func TestLifecycle_Equal(t *testing.T) {
	t.Parallel()
	resource.Require(t, resource.UnitTest)

	now := time.Now()
	nowPlus := time.Now().Add(time.Duration(1000))

	a := dbsupport.Lifecycle{
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: nil,
	}

	// Test for type difference
	b := convert.DummyEqualer{}
	assert.False(t, a.Equal(b))

	// Test CreateAt difference
	c := dbsupport.Lifecycle{
		CreatedAt: nowPlus,
		UpdatedAt: now,
		DeletedAt: nil,
	}
	assert.False(t, a.Equal(c))

	// Test UpdatedAt difference
	d := dbsupport.Lifecycle{
		CreatedAt: now,
		UpdatedAt: nowPlus,
		DeletedAt: nil,
	}
	assert.False(t, a.Equal(d))

	// Test DeletedAt (one is not nil, the other is) difference
	e := dbsupport.Lifecycle{
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &now,
	}
	assert.False(t, a.Equal(e))

	// Test DeletedAt (both are not nil) difference
	g := dbsupport.Lifecycle{
		CreatedAt: now,
		UpdatedAt: nowPlus,
		DeletedAt: &now,
	}
	h := dbsupport.Lifecycle{
		CreatedAt: now,
		UpdatedAt: nowPlus,
		DeletedAt: &nowPlus,
	}
	assert.False(t, g.Equal(h))
}
