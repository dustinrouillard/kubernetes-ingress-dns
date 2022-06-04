package traefik_watcher

import (
	"context"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/bep/debounce"

	"github.com/dustinrouillard/kubernetes-ingress-dns/pkg/util"

	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/informers/externalversions"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

type Watcher struct {
	client   versioned.Interface
	onChange func(*util.Payload)
}

var existingIngresses = []util.IngressPayload{}

func New(client versioned.Interface, onChange func(*util.Payload)) *Watcher {
	return &Watcher{
		client:   client,
		onChange: onChange,
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	factory := externalversions.NewSharedInformerFactory(w.client, time.Minute)
	ingressLister := factory.Traefik().V1alpha1().IngressRoutes().Lister()

	onChange := func() {
		payload := &util.Payload{}

		ingresses, err := ingressLister.List(labels.Everything())
		if err != nil {
			log.Println(err)
			return
		}

		for _, ingress := range ingresses {
			serviceIngressClass := util.Getenv("INGRESS_CLASS", "")

			if serviceIngressClass == "" && util.ApplicationMode != "annotation" {
				serviceIngressClass = "nginx"
			}

			if util.ApplicationMode != "class" && ingress.Annotations["dstn.to/ingress-dns"] != "true" {
				continue
			}

			ingressPayload := util.IngressPayload{
				Traefik: ingress,
				Ingress: nil,
			}
			payload.Ingresses = append(payload.Ingresses, ingressPayload)
		}

		if !reflect.DeepEqual(payload.Ingresses, existingIngresses) {
			w.onChange(payload)
			existingIngresses = payload.Ingresses
		}
	}

	debounced := debounce.New(time.Second)
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			debounced(onChange)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			debounced(onChange)
		},
		DeleteFunc: func(obj interface{}) {
			debounced(onChange)
		},
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		informer := factory.Traefik().V1alpha1().IngressRoutes().Informer()
		informer.AddEventHandler(handler)
		informer.Run(ctx.Done())
		wg.Done()
	}()

	wg.Wait()
	return nil
}
