# Cloudflare Zero-Trust Operator

Cloudflare Zero-Trust operator allow you to manage your zero-trust configuration directly from kubernetes

<!-- Version_Placeholder -->
![Version: 0.5.1](https://img.shields.io/badge/Version-0.5.1-informational?style=flat-square)
[![CRD - reference](https://img.shields.io/badge/CRD-reference-2ea44f)](https://doc.crds.dev/github.com/BojanZelic/cloudflare-zero-trust-operator)
![Unit Tests](https://github.com/BojanZelic/cloudflare-zero-trust-operator/actions/workflows/unit.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/bojanzelic/cloudflare-zero-trust-operator)](https://goreportcard.com/report/github.com/bojanzelic/cloudflare-zero-trust-operator)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/bojanzelic/cloudflare-zero-trust-operator)](https://pkg.go.dev/github.com/bojanzelic/cloudflare-zero-trust-operator)
[![codecov](https://codecov.io/gh/BojanZelic/cloudflare-zero-trust-operator/branch/main/graph/badge.svg?token=BRSGWWVA2W)](https://codecov.io/gh/BojanZelic/cloudflare-zero-trust-operator)

## Example Usage

Cloudflare Access Group
```yaml
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessGroup
metadata:
  name: accessgroup-example
  annotations:
    # (optional) default: "false"
    #   ensures that the resource isn't removed from cloudflare if the CR is deleted
    cloudflare.zelic.io/prevent-destroy: "true"
spec:
  name: my access group
  include:
    - emails:
      - testemail1@domain.com
      - testemail2@domain.com
```

```yaml
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessApplication
metadata:
  name: domain-example
  annotations:
    # (optional) default: "false"
    #   ensures that the resource isn't removed from cloudflare if the CR is deleted
    cloudflare.zelic.io/prevent-destroy: "true"
spec:
  name: my application
  domain: domain.example.com
  autoRedirectToIdentity: true
  policies: 
    - name: Allow testemail1
      decision: allow
      include:
        - emails:
          - testemail1@domain.com
```

![Example App](./docs/images/app_example.png)

## Features
Currently in Project scope
- [x] Manage Cloudflare Access Groups
- [x] Manage Cloudflare Access Applications
- [x] Manage Cloudflare Access Tokens


## Complete Example

```yaml
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessApplication
metadata:
  name: domain-example
  annotations:
    cloudflare.zelic.io/prevent-destroy: "false"
spec:
  name: my application
  domain: domain.example.com
  autoRedirectToIdentity: true
  appLauncherVisible: true
  type: self_hosted
  allowedIdps:
    - "699d98642c564d2e855e9661899b7252"
  sessionDuration: 24h
  enableBindingCookie: false
  httpOnlyCookieAttribute: true
  logoUrl: "https://www.cloudflare.com/img/logo-web-badges/cf-logo-on-white-bg.svg"
  policies: 
    - name: Allow my rules
      decision: allow
      include:
        - emails:
          - testemail1@domain.com
        - emailDomains:
          - my-domain.com
        - ipRanges:
          - "11.22.33.44/32"
        - accessGroups:
          - value: "my-access-group-id"
        - googleGroups:
          - email: my-google-group@domain.com
            identityProviderId: 00000000-0000-0000-0000-00000000000000
        - oktaGroup:
          - name: my-okta-group
            identityProviderId: 10000000-0000-0000-0000-00000000000000
```

## Advanced Usage

See some more examples of how to use the [cloudflare zero-trust operator here](./docs/Advanced_Usage.md) 

## Install

### Token Permissions

On your Cloudflare Dashboard; Create a custom API token with the following permissions:
* Access: Service Tokens:Edit
* Access: Organizations, Identity Providers, and Groups: Edit
* Access: Apps and Policies:Edit

This token will be used referenced as `CLOUDFLARE_API_TOKEN` in the secret below; 

### Install with Helm

1) Create your namespace
```
kubectl create ns zero-trust-system
```

2) Create a secret with your cloudflare credentials

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
  CLOUDFLARE_API_EMAIL: <email>
  CLOUDFLARE_API_KEY: <api_key>
  CLOUDFLARE_API_TOKEN: <api_token>
```

3) Install the helm repo
```bash
helm repo add zelic-io https://zelic-io.github.io/charts
 
helm install --namespace=zero-trust-system --set secretRef=cloudflare-creds cloudflare-zero-trust-operator zelic-io/cloudflare-zero-trust-operator
```

## Install with Kustomize

`kustomization.yaml`
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - https://github.com/BojanZelic/cloudflare-zero-trust-operator//config/default?ref=main

secretGenerator:
- name: cloudflare-creds
  literals:
    - CLOUDFLARE_API_KEY=""
    - CLOUDFLARE_API_EMAIL=""
    - CLOUDFLARE_ACCOUNT_ID=""
    - CLOUDFLARE_API_TOKEN=""
```

## Compatability

This provider's versions are able to install and manage the following versions of Kubernetes:

|                                                | v1.22 - v1.31 | 
| ---------------------------------------------- | ----- | 
| Cloudflare Zero Trust Operator v0.0.1-current  | ✓     |


## Local Development

```
cp .env.example .env.integration
vim .env.integration # add your creds here
make integration-test
```

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

