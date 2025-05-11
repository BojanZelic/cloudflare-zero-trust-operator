# Advanced Usage

## Regarding Identity Provider IDs
You might need to specify Identity Provider IDs in your specifications; to know which are available to you and configured on your CloudFlare's account, you might find a recap of those in the log when you start your operator.
```
// TODO add log example
```

## Reference other resources from another namespace

We can reference secondary resources (eg `CloudflareServiceToken`, `CloudflareAccessGroup`, `CloudflareAccessReusablePolicy`) directly from another namespace. Just ensure RBAC permissions are set accordingly.

ex:

Cloudflare Access Group
```yaml
apiVersion: cloudflare.zelic.io/v4alpha1
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

Cloudflare Reusable Policy
```yaml
apiVersion: cloudflare.zelic.io/v4alpha1
kind: CloudflareAccessReusablePolicy
metadata:
  name: allow-testemail1
  namespace: default
spec:
  name: Allow testemail1
  decision: allow
  include:
    - emails:
      - testemail3@domain.com
    - groupRefs:
      - accessgroup-example # or, use "{metadata.namespace}/{metadata.name}" for explicit targeting
```

Cloudflare Application
```yaml
apiVersion: cloudflare.zelic.io/v4alpha1
kind: CloudflareAccessApplication
metadata:
  name: domain-example
  namespace: default2 # another namespace here
spec:
  name: my application
  domain: domain.example.com
  # type: self_hosted (implicit)
  autoRedirectToIdentity: true
  policyRefs:
    - default/allow-testemail1 # access resource from "default" instead of "default2"
```

## Regarding `WARP` and `App Launcher` special applications

In Cloudflare's Zero Trust Official Dashboard UI, both `WARP`'s Device enrollment permissions and `App Launcher` are not explicitely considered Access Applications per-se; You cannot find them in `Zero Trust > [YourAccount] > Access > Applications`.

But, considering them with CloudFlare's backend / API logic, they are ! These are special cases of Access Applications, in a sense that both types are to be unique per CloudFlare Account.

As such, you can still use a `CloudflareAccessApplication` to configure their policies. Just make sure to not declare any of those multiple times in your cluster, and to activate any of these functionalities accordingly beforehand in your CloudFlare's Zero Trust Dashboard.

ex:

For WARP's Device Enrollment Permissions
```yaml
apiVersion: cloudflare.zelic.io/v4alpha1
kind: CloudflareAccessGroup
metadata:
  name: unique_warp_dep_app
  namespace: default
spec:
  type: warp
  # name: Warp Login App # NOT NEEDED ! "Warp Login App" is the default name (per CloudFlare's API) and cannot be changed
  # domain: "whatever.com" # NOT NEEDED ! Meaningless in this context
  # policyRefs: ....
```

For App Launcher
```yaml
apiVersion: cloudflare.zelic.io/v4alpha1
kind: CloudflareAccessApplication
metadata:
  name: unique_app_launcher_app
  namespace: default
spec:
  type: app_launcher
  # name: App Launcher # NOT NEEDED ! "App Launcher" is the default name (per CloudFlare's API) and cannot be changed
  # domain: "whatever.com" # NOT NEEDED ! Meaningless in this context
  # policyRefs: ....
```