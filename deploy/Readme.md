# Deploying loki-rule-operator

This folder contains an example how the loki-rule-operator can be deployed on Kubernetes.
The deployment contains a `ValidatingWebhookConfiguration`, which validates the `LokiRules` when they are created or updated in the Kubernetes Api. This is helpfull, but not mandatory. This example uses cert-manager for creating and injecting a self-signed certificate for the webhook.

## Step 1: Replace all variables in the manifests:
Replace the following variables in the manifests:
* `TAG`: image tag of the loki-rule-operator.
* `LOKI_RULES_CONFIGMAP_NAME`: Name of the ConfigMap where all the rules will be stored. This ConfigMap should be mounted into the Loki Pod.
* `LOKI_RULES_CONFIGMAP_NAMESPACE`: Namespace of the ConfigMap where all the rules will be stored.

## Step 2: Create the CustomResourceDefintions (CRD)
```shell
kubectl apply -f crd
```

## Step 3: Deploy the operator
```shell
kubectl apply -f .
```
After the operator is created, the LOKI_RULES_CONFIGMAP

## Step 4: Mount rules configmap into the Loki pod

Configure the Ruler in the `loki.yml`:
```yaml
ruler:
  storage:
    type: local
    local:
      directory: /etc/loki/rules
```

Alter the deployment of Loki to mount the LOKI_RULES_CONFIGMAP.
```yaml
...
spec:
  ...
  template:
    spec:
      containers:
      - name: loki
        ...
        volumeMounts:
        - name: loki-rules
          mountPath: /etc/loki/rules/fake
      ...
      volumes:
      - name: loki-rules
        configMap:
          defaultMode: 420
          name: {{ $LOKI_RULES_CONFIGMAP_NAME }}
```