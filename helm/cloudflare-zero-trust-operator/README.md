# cloudflare-zero-trust-operator

![Version: 0.7.1](https://img.shields.io/badge/Version-0.7.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.7.1](https://img.shields.io/badge/AppVersion-0.7.1-informational?style=flat-square)

Operator for managing Cloudflare Zero Trust settings

Cloudflare Zero-Trust operator allow you to manage your zero-trust configuration directly from kubernetes.

**Homepage:** <https://github.com/bojanzelic/cloudflare-zero-trust-operator>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| cloudflare_account_id | string | `""` | Cloudflare Account ID - required (or set secretRef) |
| cloudflare_api_email | string | `""` | Cloudflare API Email - required (one of cloudflare_api_token or cloudflare_api_key + cloudflare_api_email) (or set secretRef) |
| cloudflare_api_key | string | `""` | API Key from cloudflare - required (one of cloudflare_api_token or cloudflare_api_key + cloudflare_api_email) (or set secretRef) |
| cloudflare_api_token | string | `""` | Cloudflare API Token - required (one of cloudflare_api_token or cloudflare_api_key + cloudflare_api_email) (or set secretRef) |
| fullnameOverride | string | `""` | override name for helm chart |
| image.pullPolicy | string | `"IfNotPresent"` | manager pullPolicy |
| image.repository | string | `"ghcr.io/bojanzelic/cloudflare-zero-trust-operator"` | manager image repo |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | config reference for pulling containers |
| manager.resources | object | `{}` | limits & requests(cpu & memory) to apply to the manager container |
| nameOverride | string | `""` | override name for helm chart |
| podAnnotations | object | `{}` | annotations to add to the pod |
| proxy.resources | object | `{}` | limits & requests(cpu & memory) to apply to the manager container |
| replicaCount | int | `1` | number of replicas to run |
| secretRef | string | `""` | name of the secret that contains the following keys: CLOUDFLARE_ACCOUNT_ID, CLOUDFLARE_API_KEY, CLOUDFLARE_API_EMAIL, CLOUDFLARE_API_TOKEN |
| service.port | int | `8443` | port of service |
| service.type | string | `"ClusterIP"` | type of service |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated using the fullname template |

## Installing

### Install with Helm

1) Create your namespace
```
kubectl create ns zero-trust-system
```

2) Create a secret with your cloudflare credentials (Alternatively these values can be supplied via values.yaml)

```yaml
apiVersion: v1
metadata:
  name: cloudflare-creds
  namespace: zero-trust-system
kind: Secret
type: Opaque
stringData:
  CLOUDFLARE_ACCOUNT_ID: <id>
  # Either EMAIL+KEY or TOKEN must be supplied
  # note: keys must still be defined even if they are empty
  CLOUDFLARE_API_EMAIL: <email>
  CLOUDFLARE_API_KEY: <api_key>
  CLOUDFLARE_API_TOKEN: <api_token>
```

3) Install the helm repo

```bash
helm repo add zelic-io https://zelic-io.github.io/charts

helm install --namespace=zero-trust-system --set secretRef=cloudflare-creds cloudflare-zero-trust-operator zelic-io/cloudflare-zero-trust-operator
```

---
