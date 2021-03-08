/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	loggingv1beta1 "github.com/opsgy/loki-rule-operator/api/v1beta1"
	"github.com/opsgy/loki-rule-operator/controllers"
	// +kubebuilder:scaffold:imports
)

// Created so that multiple inputs can be accecpted
type labelFlags []controllers.Label

func (l *labelFlags) String() string {
	str := ""
	for i, label := range *l {
		if i > 0 {
			str += ","
		}
		str += label.Name + "=" + label.Value
	}
	// change this, this is just can example to satisfy the interface
	return str
}

func (l *labelFlags) Set(value string) error {
	pair := strings.Split(value, "=")
	if len(pair) != 2 {
		return fmt.Errorf("invalid label, should be in the format <key>=<value>")
	}
	label := controllers.Label{
		Name:  strings.TrimSpace(pair[0]),
		Value: strings.TrimSpace(pair[1]),
	}
	*l = append(*l, label)
	return nil
}

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(loggingv1beta1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var rulesCM string
	var enableWebhook bool
	var externalLabels labelFlags
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&rulesCM, "rules-configmap", "default/loki-rules", "Configmap name to store all the LokiRules, in the format '<namespace>/<name>'")
	flag.BoolVar(&enableWebhook, "enable-webhook", false, "Enable validation webhook")
	flag.Var(&externalLabels, "external-label", "Add labels to the alert rules")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	config := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "934c0416.opsgy.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	rulesCMParts := strings.Split(rulesCM, "/")
	if len(rulesCMParts) != 2 {
		setupLog.Error(err, "invalid value for --rules-configmap")
		os.Exit(1)
	}

	// Create cm if not exists
	_, err = clientset.CoreV1().ConfigMaps(rulesCMParts[0]).Get(context.TODO(), rulesCMParts[1], metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			labelMap := make(map[string]string)
			labelMap["app.kubernetes.io/managed-by"] = "loki-rule-operator"
			dataMap := make(map[string]string)

			cm := &v1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      rulesCMParts[1],
					Namespace: rulesCMParts[0],
					Labels:    labelMap,
				},
				Data: dataMap,
			}
			_, err := clientset.CoreV1().ConfigMaps(rulesCMParts[0]).Create(context.TODO(), cm, metav1.CreateOptions{})
			if err != nil {
				setupLog.Error(err, "Failed to create configmap")
				os.Exit(1)
			}
		} else {
			setupLog.Error(err, "Failed to get configmap")
			os.Exit(1)
		}
	}

	if err = (&controllers.LokiRuleReconciler{
		Client:                  mgr.GetClient(),
		Log:                     ctrl.Log.WithName("controllers").WithName("LokiRule"),
		Scheme:                  mgr.GetScheme(),
		Clientset:               clientset,
		RulesConfigMapName:      rulesCMParts[1],
		RulesConfigMapNamespace: rulesCMParts[0],
		ExternalLabels:          externalLabels,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LokiRule")
		os.Exit(1)
	}
	if err = (&controllers.GlobalLokiRuleReconciler{
		Client:                  mgr.GetClient(),
		Log:                     ctrl.Log.WithName("controllers").WithName("GlobalLokiRule"),
		Scheme:                  mgr.GetScheme(),
		Clientset:               clientset,
		RulesConfigMapName:      rulesCMParts[1],
		RulesConfigMapNamespace: rulesCMParts[0],
		ExternalLabels:          externalLabels,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GlobalLokiRule")
		os.Exit(1)
	}
	if enableWebhook {
		if err = (&loggingv1beta1.LokiRule{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "LokiRule")
			os.Exit(1)
		}
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
