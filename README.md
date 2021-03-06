# Kubernetes Ingress DNS

This is a kubernetes deployment which will monitor ingress resources, and crd ingresses from traefik, and populate the host mappings on a dnsmasq server hosted internally.

I built this for accessing internal services without any complication from outside the cluster, and through my wireguard vpn.

Currently this only works with ingresses that use a LoadBalancer service, as it relies on the `status.loadBalancer.Ingress.IP`

## Deployment

The manifests for deploying are stored in [manifests](manifests) so you should make any changes you need to make, most noteablly the environment variables, then just simple apply them, making sure the `rbac.yaml` is applied first.

**Note: this puts everything in the default namespace, if you want to change that just change all occurrences of `namespace: default` in the manifests.**

## Customizing dnsmasq

The dnsmasq config is mounted from `/etc/dnsmasq/dnsmasq.conf` so you can overrwite this with a configmap or volume if you need to make any modifications to the way the configuration works, such as turning on `no-resolv` and setting your own upstream, or something along those lines.

## Environment Variables

| Name                      | Options                  | Default                        | Description                                                                                |
| ------------------------- | ------------------------ | ------------------------------ | ------------------------------------------------------------------------------------------ |
| INGRESS_SERVICE_NAMESPACE |                          | None **Required**              | The namespace that contains the LoadBalancer service for your ingress controller           |
| INGRESS_SERVICE_NAME      |                          | None **Required**              | The name of the LoadBalancer service used for routing of your ingress controller           |
| INGRESS_PROVIDER          | `ingress`, `traefik`     | ingress _Optional_             | Configures what resources the controller watches for ingress hosts                         |
| MODE                      | `all, annotation`        | `annotation`                   | Mode in which the ingress watcher runs in                                                  |
| HOSTS_PATH                |                          | `/hosts` (Make sure exists)    | The directory (which must match from dnsmasq.conf) that ingress hosts files will be put in |

## Monitoring Ingresses

There is a couple ways to monitor ingress resources, and they can all be used at once if needed.

You can target by annotation, or no filter at the moment, planning on more filtering in the future, with labels, ingress class, ect.

### Annotations

To track ingress resources based on annotations add the following to your ingress resource.

```yml
metadata:
  annotations:
    dstn.to/ingress-dns: "true"
```
