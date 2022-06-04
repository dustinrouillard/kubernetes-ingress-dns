package ingress_watcher

import (
	"context"
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/bep/debounce"

	"github.com/dustinrouillard/kubernetes-ingress-dns/pkg/util"

	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Watcher struct {
	client   kubernetes.Interface
	onChange func(*util.Payload)
}

var existingIngresses = []util.IngressPayload{}

func New(client kubernetes.Interface, onChange func(*util.Payload)) *Watcher {
	return &Watcher{
		client:   client,
		onChange: onChange,
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	factory := informers.NewSharedInformerFactory(w.client, time.Minute)
	serviceLister := factory.Core().V1().Services().Lister()
	ingressLister := factory.Networking().V1().Ingresses().Lister()

	addBackend := func(ingressPayload *util.IngressPayload, backend networking.IngressBackend) {
		svc, err := serviceLister.Services(ingressPayload.Ingress.Namespace).Get(backend.Service.Name)
		if err != nil {
			log.Println(err)
		} else {
			m := make(map[string]int)
			for _, port := range svc.Spec.Ports {
				m[port.Name] = int(port.Port)
			}
			ingressPayload.ServicePorts[svc.Name] = m
		}
	}

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

			if util.ApplicationMode == "class" && ingress.Spec.IngressClassName != &serviceIngressClass {
				continue
			} else if util.ApplicationMode == "annotation" && ingress.Annotations["dstn.to/ingress-dns"] != "true" {
				continue
			}

			ingressPayload := util.IngressPayload{
				Ingress:      ingress,
				Traefik:      nil,
				ServicePorts: make(map[string]map[string]int),
			}
			payload.Ingresses = append(payload.Ingresses, ingressPayload)

			if ingress.Spec.DefaultBackend != nil {
				addBackend(&ingressPayload, *ingress.Spec.DefaultBackend)
			}

			for _, rule := range ingress.Spec.Rules {
				if rule.HTTP != nil {
					continue
				}

				for _, path := range rule.HTTP.Paths {
					addBackend(&ingressPayload, path.Backend)
				}
			}
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
		informer := factory.Core().V1().Secrets().Informer()
		informer.AddEventHandler(handler)
		informer.Run(ctx.Done())
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		informer := factory.Networking().V1().Ingresses().Informer()
		informer.AddEventHandler(handler)
		informer.Run(ctx.Done())
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		informer := factory.Core().V1().Services().Informer()
		informer.AddEventHandler(handler)
		informer.Run(ctx.Done())
		wg.Done()
	}()

	wg.Wait()
	return nil
}
