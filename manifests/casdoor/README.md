# Casdoor

[Casdoor](https://casdoor.ai) is an open-source Identity and Access Management (IAM) / single sign-on platform with a web UI, supporting OAuth 2.0, OIDC, SAML, CAS, LDAP, SCIM, WebAuthn, TOTP and MFA. This is the official Helm chart, published with every Casdoor release.

- Documentation: <https://casdoor.ai/docs/basic/try-with-helm>
- Source: <https://github.com/casdoor/casdoor/tree/master/manifests/casdoor>
- Demo: <https://door.casdoor.com>

## Prerequisites

- Kubernetes 1.19+
- Helm v3.8+

## Installation

```shell
helm install casdoor oci://ghcr.io/casdoor/helm-charts/casdoor --version <version>
```

To install with a custom values file:

```shell
helm install casdoor oci://ghcr.io/casdoor/helm-charts/casdoor \
  --version <version> \
  -f my-values.yaml
```

The chart does not expose Casdoor outside the cluster by default. Use `kubectl port-forward svc/<service-name> 8000:8000` to try it, or enable Ingress or the Gateway API below.

## Upgrading

```shell
helm upgrade casdoor oci://ghcr.io/casdoor/helm-charts/casdoor --version <version>
```

Charts up to 4.15.0 were published as `oci://registry-1.docker.io/casbin/casdoor-helm-charts`, which still receives every release. Running the `helm upgrade` command above moves an existing release to this chart; resource names stay the same.

## Uninstalling

```shell
helm uninstall casdoor
```

## Configuration

Override any value from [values.yaml](https://github.com/casdoor/casdoor/blob/master/manifests/casdoor/values.yaml) using `--set` or a custom values file.

### Core parameters

| Parameter | Description | Default |
|---|---|---|
| `replicaCount` | Number of Casdoor pods | `1` |
| `image.repository` | Docker image repository | `casbin` |
| `image.name` | Docker image name | `casdoor` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `image.tag` | Image tag (defaults to chart appVersion) | `""` |
| `config` | Casdoor `app.conf` content (Go template) | See values.yaml |
| `configFromSecret` | Name of an existing Secret holding `app.conf` | `""` |

### Database

| Parameter | Description | Default |
|---|---|---|
| `database.driver` | Database driver: `mysql`, `postgres`, `cockroachdb`, `sqlite` | `sqlite` |
| `database.user` | Database username | `""` |
| `database.password` | Database password | `""` |
| `database.host` | Database host | `""` |
| `database.port` | Database port (empty = driver default) | `""` |
| `database.databaseName` | Database name | `casdoor` |
| `database.sslMode` | SSL mode for the connection | `disable` |

### LDAP

| Parameter | Description | Default |
|---|---|---|
| `ldap.enabled` | Enable the built-in LDAP server | `false` |
| `ldap.service.port` | LDAP service port | `389` |

### Declarative configuration (init data)

Organizations, applications, users, providers, roles, permissions and the other Casdoor objects can be kept in the values file, in the [init data](https://casdoor.ai/docs/deployment/data-initialization) format. The chart stores them in a Secret, Casdoor applies them at startup and checks them for changes every `initData.watchInterval` seconds, so a `helm upgrade` that changes them takes effect without restarting the pods.

```yaml
initData:
  enabled: true
  data:
    organizations:
      - owner: admin
        name: acme
        displayName: Acme
        passwordType: bcrypt
    applications:
      - owner: admin
        name: app-acme
        organization: acme
        displayName: Acme Portal
        redirectUris:
          - https://portal.acme.example.com/callback
    users:
      - owner: acme
        name: alice
        displayName: Alice
        password: change-me
        signupApplication: app-acme
```

With `initData.merge: true` (the default), an object that already exists only gets the fields written in the data, its other fields keep the values edited in the web UI. A user's `password` is only used when the user is created, so users can change their own passwords. Objects removed from the data are not deleted from Casdoor.

| Parameter | Description | Default |
|---|---|---|
| `initData.enabled` | Apply `initData.data` (or `initData.existingSecret`) | `false` |
| `initData.merge` | Update existing objects with the given fields only; when `false`, they are deleted and re-created on every apply | `true` |
| `initData.watchInterval` | Seconds between the checks for changes, `0` applies the data only at startup (the pods then restart when `initData.data` changes) | `30` |
| `initData.existingSecret` | Name of an existing Secret holding the data, instead of `initData.data` | `""` |
| `initData.existingSecretKey` | Key of the file in `existingSecret`; a `.yaml`/`.yml` key is read as YAML, any other as JSON | `init_data.yaml` |
| `initData.data` | The objects to apply | `{}` |

### Service

| Parameter | Description | Default |
|---|---|---|
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8000` |
| `service.annotations` | Annotations to add to the Service | `{}` |

### Ingress

Exposes Casdoor via a standard Kubernetes `Ingress` resource.

| Parameter | Description | Default |
|---|---|---|
| `ingress.enabled` | Enable Ingress | `false` |
| `ingress.className` | IngressClass name | `""` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `ingress.hosts` | List of `{host, paths}` entries | See values.yaml |
| `ingress.tls` | TLS configuration | `[]` |

Example:

```yaml
ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: casdoor.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: casdoor-tls
      hosts:
        - casdoor.example.com
```

### Gateway API

Exposes Casdoor via the [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/) (`HTTPRoute` + optional `Gateway`). This is the modern successor to Ingress, supported by Istio, Envoy Gateway, Cilium, Kong, NGINX Gateway Fabric, and others.

**Requires:** Gateway API CRDs (`gateway.networking.k8s.io/v1`) installed in the cluster and a compatible controller running.

#### Core Gateway API parameters

| Parameter | Description | Default |
|---|---|---|
| `gatewayApi.enabled` | Enable HTTPRoute creation | `false` |
| `gatewayApi.createGateway` | Also create a `Gateway` resource | `false` |
| `gatewayApi.annotations` | Annotations for the HTTPRoute | `{}` |
| `gatewayApi.labels` | Extra labels for the HTTPRoute | `{}` |
| `gatewayApi.parentRefs` | Parent Gateway references for the HTTPRoute | `[]` |
| `gatewayApi.hostnames` | Hostnames to match (Host header) | `[]` |
| `gatewayApi.rules` | Routing rules (matches, filters, backendRefs) | PathPrefix `/` → chart Service |

#### Gateway creation parameters (`gatewayApi.gateway.*`)

Only used when `gatewayApi.createGateway=true`.

| Parameter | Description | Default |
|---|---|---|
| `gatewayApi.gateway.name` | Gateway resource name (defaults to chart fullname) | `""` |
| `gatewayApi.gateway.gatewayClassName` | **Required.** GatewayClass name (e.g. `istio`, `envoy`) | `""` |
| `gatewayApi.gateway.annotations` | Annotations for the Gateway | `{}` |
| `gatewayApi.gateway.labels` | Extra labels for the Gateway | `{}` |
| `gatewayApi.gateway.listeners` | Gateway listeners | HTTP:80 (see values.yaml) |

#### HTTPS redirect parameters (`gatewayApi.httpsRedirect.*`)

Creates a second HTTPRoute that redirects HTTP traffic to HTTPS.

| Parameter | Description | Default |
|---|---|---|
| `gatewayApi.httpsRedirect.enabled` | Enable the HTTP→HTTPS redirect HTTPRoute | `false` |
| `gatewayApi.httpsRedirect.httpListenerName` | HTTP listener `sectionName` on the Gateway | `http` |
| `gatewayApi.httpsRedirect.httpsListenerName` | HTTPS listener `sectionName` on the Gateway | `https` |
| `gatewayApi.httpsRedirect.hostnames` | Hostnames for the redirect route (falls back to `gatewayApi.hostnames`) | `[]` |
| `gatewayApi.httpsRedirect.statusCode` | Redirect response code | `301` |
| `gatewayApi.httpsRedirect.parentRefs` | Override parentRefs for the redirect route | `[]` |

#### Example: attach to an existing Gateway

```yaml
gatewayApi:
  enabled: true
  parentRefs:
    - name: my-gateway
      namespace: gateway-system
      sectionName: https
  hostnames:
    - casdoor.example.com
```

#### Example: create a new Gateway (Istio) with HTTP→HTTPS redirect

```yaml
gatewayApi:
  enabled: true
  createGateway: true
  hostnames:
    - casdoor.example.com
  gateway:
    gatewayClassName: istio
    listeners:
      - name: http
        protocol: HTTP
        port: 80
        allowedRoutes:
          namespaces:
            from: Same
      - name: https
        protocol: HTTPS
        port: 443
        tls:
          certificateRefs:
            - name: casdoor-tls
              kind: Secret
        allowedRoutes:
          namespaces:
            from: Same
  httpsRedirect:
    enabled: true
```

### Resources and scaling

| Parameter | Description | Default |
|---|---|---|
| `resources` | CPU/memory requests and limits | `{}` |
| `autoscaling.enabled` | Enable HorizontalPodAutoscaler | `false` |
| `autoscaling.minReplicas` | HPA minimum replicas | `1` |
| `autoscaling.maxReplicas` | HPA maximum replicas | `100` |
| `autoscaling.targetCPUUtilizationPercentage` | HPA CPU target | `80` |

### Scheduling

| Parameter | Description | Default |
|---|---|---|
| `nodeSelector` | Node selector labels | `{}` |
| `tolerations` | Pod tolerations | `[]` |
| `affinity` | Pod affinity/anti-affinity rules | `{}` |
| `priorityClassName` | Pod priority class | `""` |

### Extra containers and volumes

| Parameter | Description | Default |
|---|---|---|
| `initContainersEnabled` | Enable init containers | `false` |
| `initContainers` | Init container definitions (YAML string) | `""` |
| `extraContainersEnabled` | Enable sidecar containers | `false` |
| `extraContainers` | Sidecar container definitions (YAML string) | `""` |
| `extraVolumeMounts` | Extra volume mounts for the Casdoor container | `[]` |
| `extraVolumes` | Extra volumes for the pod | `[]` |
| `defaultConfigVolumeEnabled` | Mount the default config volume at `/conf` | `true` |

### Environment variables

| Parameter | Description | Default |
|---|---|---|
| `envFromSecret` | Env vars sourced from individual Secret keys | `[]` |
| `envFromConfigmap` | Env vars sourced from individual ConfigMap keys | `[]` |
| `envFrom` | Env vars from entire Secrets or ConfigMaps | `[]` |

### Probes

| Parameter | Description | Default |
|---|---|---|
| `probe.readiness.enabled` | Enable readiness probe | `true` |
| `probe.liveness.enabled` | Enable liveness probe | `true` |
