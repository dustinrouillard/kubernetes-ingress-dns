package util

import (
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"

	networking "k8s.io/api/networking/v1"
)

type IngressPayload struct {
	Ingress      *networking.Ingress
	Traefik      *v1alpha1.IngressRoute
	ServicePorts map[string]map[string]int
}

type Payload struct {
	Ingresses []IngressPayload
}
