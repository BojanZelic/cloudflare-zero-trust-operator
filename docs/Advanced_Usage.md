# Advanced Usage

## Reference other resources

We can reference other resources (ex: CloudflareServiceToken or CloudflareAccessGroup) directly from an CloudflareAccessApplicaion

ex:

Cloudflare Access Group
```yaml
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessGroup
metadata:
  name: accessgroup-example
  namespace: default
spec:
  name: my access group
  include:
    - emails:
      - testemail1@domain.com
      - testemail2@domain.com
```

Cloudflare Application
```yaml
apiVersion: cloudflare.zelic.io/v1alpha1
kind: CloudflareAccessApplication
metadata:
  name: domain-example
  namespace: default
spec:
  name: my application
  domain: domain.example.com
  autoRedirectToIdentity: true
  policies: 
    - name: Allow testemail1
      decision: allow
      include:
        - emails:
          - testemail3@domain.com
        - accessGroups:
            - valueFrom:
                name: accessgroup-example
                namespace: default
```