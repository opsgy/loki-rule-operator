# loki-rule-operator
Kubernetes Operator that adds the `LokiRule` and `GlobalLokiRule` Custom Resource Definitions to a cluster. These resources helps configuring Alert rules for your Loki setup.

The LogQL expressions are evaluated before the rules are applied.

## Example
```yaml
apiVersion: logging.opsgy.com/v1beta1
kind: LokiRule
metadata:
  name: credentials-leak
  namespace: prod
spec:
  groups:
  - name: credentials_leak
    rules:
    - alert: http-credentials-leaked
      annotations:
        message: '{{ $labels.job }} is leaking http basic auth credentials.'
      expr: sum by (cluster, job, pod) (count_over_time({namespace="prod"} |~ "http(s?)://(\\w+):(\\w+)@"
        [5m]) > 0)
      for: 10m
      labels:
        severity: critical
```

## Difference between cluster-wide and namespaced
Namespaced rules will be enforece to have the selector `namespace` set to the namespace where the rule is created in.

## Setup the loki-rule-operator
1. Add the crd's from `config/crd/bases`
2. Deploy the operator, an example can be found in the `deploy` folder

## license
MIT