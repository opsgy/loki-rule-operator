# loki-rule-operator
Loki-rule-operator is light-weight Kubernetes Operator that adds the `LokiRule` and `GlobalLokiRule` Custom Resource Definitions to a cluster. These resources helps configuring [Alert rules](https://grafana.com/docs/loki/latest/alerting/) for your [Loki setup](https://grafana.com/docs/loki/latest/installation/).

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
      expr: sum by (cluster, job, pod) (count_over_time({namespace="prod"} |~ "http(s?)://(\\w+):(\\w+)@" [5m]) > 0)
      for: 10m
      labels:
        severity: critical
```

## Difference between `GlobalLokiRule` and `LokiRule`
`LokiRule` is a namespaced resource and will will enforce the selector `{namespace="<namespace>"}` on the LogQL expression. The `GlobalLokiRule` is cluster wide and doesn't enforce the namespace selector.

## Setup the loki-rule-operator
See the [deploy](./deploy) folder.

## License
Apache License 2.0, see [LICENSE](LICENSE).