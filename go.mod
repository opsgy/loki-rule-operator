module github.com/opsgy/loki-rule-operator

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/grafana/loki v1.6.1
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/prometheus/prometheus v1.8.2-0.20201119181812-c8f810083d3f
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v12.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.7.0
)

replace k8s.io/client-go => k8s.io/client-go v0.19.4

// Hack to use the latest version of loki
replace github.com/grafana/loki v1.6.1 => ../../../../src/lib/github.com/grafana/loki
