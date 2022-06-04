package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustinrouillard/kubernetes-ingress-dns/pkg/util"
	IngressWatcher "github.com/dustinrouillard/kubernetes-ingress-dns/pkg/watchers/ingress"
	TraefikWatcher "github.com/dustinrouillard/kubernetes-ingress-dns/pkg/watchers/traefik"

	TraefikHttp "github.com/traefik/traefik/v2/pkg/muxer/http"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"

	"golang.org/x/sync/errgroup"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	client, err := kubernetes.NewForConfig(KubernetesConfig())
	if err != nil {
		log.Println(err)
	}

	provider := util.Getenv("INGRESS_PROVIDER", "ingress")

	if provider == "ingress" {
		ingress := IngressWatcher.New(client, func(payload *util.Payload) {
			_, ctx := errgroup.WithContext(context.Background())

			lines := []string{}

			for _, ingress := range payload.Ingresses {
				serviceNameSpace := os.Getenv("INGRESS_SERVICE_NAMESPACE")
				serviceName := os.Getenv("INGRESS_SERVICE_NAME")

				ingressService, err := client.CoreV1().Services(serviceNameSpace).Get(ctx, serviceName, v1.GetOptions{})

				if err != nil {
					log.Println(err)
				}

				ingressLoadbalancerIp := ingressService.Status.LoadBalancer.Ingress[len(ingressService.Status.LoadBalancer.Ingress)-1].IP

				hosts := []string{}

				for _, rule := range ingress.Ingress.Spec.Rules {
					hosts = append(hosts, rule.Host)
				}

				lines = append(lines, ingressLoadbalancerIp+" "+strings.Join(hosts, " "))
			}

			util.WriteHosts(lines, "ingress.hosts")
		})

		ingressErrorGroup, ctx := errgroup.WithContext(context.Background())

		ingressErrorGroup.Go(func() error {
			return ingress.Run(ctx)
		})

		if err := ingressErrorGroup.Wait(); err != nil {
			log.Println(err)
		}
	} else if provider == "traefik" {
		traefikClient, traefikErr := versioned.NewForConfig(KubernetesConfig())
		if traefikErr != nil {
			log.Println(traefikErr)
		}

		traefik := TraefikWatcher.New(traefikClient, func(payload *util.Payload) {
			_, ctx := errgroup.WithContext(context.Background())

			lines := []string{}

			for _, ingress := range payload.Ingresses {
				serviceNameSpace := os.Getenv("INGRESS_SERVICE_NAMESPACE")
				serviceName := os.Getenv("INGRESS_SERVICE_NAME")

				ingressService, err := client.CoreV1().Services(serviceNameSpace).Get(ctx, serviceName, v1.GetOptions{})

				if err != nil {
					log.Println(err)
				}

				ingressLoadbalancerIp := ingressService.Status.LoadBalancer.Ingress[len(ingressService.Status.LoadBalancer.Ingress)-1].IP

				hosts := []string{}

				for _, rule := range ingress.Traefik.Spec.Routes {
					parsed, err := TraefikHttp.ParseDomains(rule.Match)

					if err != nil {
						log.Println(err)
					}

					if parsed == nil {
						continue
					}

					hosts = append(hosts, parsed[len(parsed)-1])
				}

				if len(hosts) > 0 {
					lines = append(lines, ingressLoadbalancerIp+" "+strings.Join(hosts, " "))
				}
			}

			util.WriteHosts(lines, "traefik.hosts")
		})

		treafikErrorGroup, ctx := errgroup.WithContext(context.Background())

		treafikErrorGroup.Go(func() error {
			return traefik.Run(ctx)
		})

		if err := treafikErrorGroup.Wait(); err != nil {
			log.Println(err)
		}
	}
}

func KubernetesConfig() *rest.Config {
	config, err := rest.InClusterConfig()

	if err != nil {
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	}

	if err != nil {
		log.Println(err)
	}

	return config
}
