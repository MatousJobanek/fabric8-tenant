package openshift

import (
	"sync"
	"github.com/fabric8-services/fabric8-tenant/template"
	"net/http"
	"sort"
	"github.com/fabric8-services/fabric8-tenant/retry"
	"time"
	"fmt"
	"github.com/prometheus/common/log"
)

func processAndApplyAll(builder *ClientWithObjectsBuilder, action string) error {
	var templatesWait sync.WaitGroup
	templatesWait.Add(len(builder.templates))

	for _, tmpl := range builder.templates {
		go processAndApply(&templatesWait, tmpl, template.CollectVars(builder.user, builder.config), *builder.client, action)
	}
	templatesWait.Wait()
	return nil
}

func processAndApply(templatesWait *sync.WaitGroup, tmpl template.Template, vars map[string]string, client Client, action string) {
	defer templatesWait.Done()
	objects, err := tmpl.Process(vars)

	if err != nil {
		client.Log.Error(err)
		return
	}
	if action == http.MethodDelete {
		sort.Reverse(template.ByKind(objects))
	} else {
		sort.Sort(template.ByKind(objects))
	}

	var objectsWait sync.WaitGroup
	objectsWait.Add(len(objects))

	for _, object := range objects {
		//fmt.Println("created for objects")
		go apply(&objectsWait, client, action, object)
	}
	objectsWait.Wait()
}

func apply(objectsWait *sync.WaitGroup, client Client, action string, object template.Object) {
	defer objectsWait.Done()
	errs := retry.Do(5, time.Millisecond*50, func() error {
		objectEndpoint, found := objectEndpoints[template.GetKind(object)]
		//fmt.Println(template.GetKind(object))
		if !found {
			return fmt.Errorf("there is no supported endpoint for the object %s", template.GetKind(object))

		}
		_, err := objectEndpoint.Apply(&client, object, action)
		return err
	})
	if len(errs) != 0 {
		log.Error(errs)
	}
}
