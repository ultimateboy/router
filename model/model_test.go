package model

import (
	"reflect"
	"testing"

	"k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
)

const (
	routerName      = "deis-router"
	routerNamespace = "deis"
)

func TestMergeRouterConfig(t *testing.T) {
	replicas := int32(1)
	routerDeployment := v1beta1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      routerName,
			Namespace: routerNamespace,
			Annotations: map[string]string{
				"router.deis.io/nginx.defaultTimeout":             "1500s",
				"router.deis.io/nginx.ssl.bufferSize":             "6k",
				"router.deis.io/nginx.ssl.hsts.maxAge":            "1234",
				"router.deis.io/nginx.ssl.hsts.includeSubDomains": "true",
			},
		},
		Spec: v1beta1.DeploymentSpec{
			Strategy: v1beta1.DeploymentStrategy{
				Type:          v1beta1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &v1beta1.RollingUpdateDeployment{},
			},
			Replicas: &replicas,
			Selector: &v1beta1.LabelSelector{MatchLabels: map[string]string{"app": routerName}},
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": routerName,
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: "deis/router",
						},
					},
				},
			},
		},
	}

	routerConfigMap := v1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      routerName,
			Namespace: routerNamespace,
		},
		Data: map[string]string{
			"nginx.errorLogLevel":              "info",
			"nginx.ssl.bufferSize":             "2k",
			"nginx.ssl.hsts.preload":           "true",
			"nginx.ssl.hsts.includeSubDomains": "false",
		},
	}

	expectedConfig := newRouterConfig()
	sslConfig := newSSLConfig()
	hstsConfig := newHSTSConfig()

	// A value not set in the deployment annotations or the ConfigMap (should be default value.)
	expectedConfig.MaxWorkerConnections = "768"
	// A value set only in the deployment annotations.
	expectedConfig.DefaultTimeout = "1500s"
	// A value set only in the configmap.
	expectedConfig.ErrorLogLevel = "info"

	// A nested value set in both the deployment annotations and the configmap.
	sslConfig.BufferSize = "6k"

	// A nested+nested value set only in the deployment annotations.
	hstsConfig.MaxAge = 1234
	// A nested+nested value set only in the configmap.
	hstsConfig.Preload = true
	// A nested+nested value set in both the deployment and the configmap.
	hstsConfig.IncludeSubDomains = true

	sslConfig.HSTSConfig = hstsConfig
	expectedConfig.SSLConfig = sslConfig

	actualConfig, err := mergeRouterConfig(&routerDeployment, &routerConfigMap)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(expectedConfig, actualConfig) {
		t.Errorf("Expected routerConfig does not match actual.")

		t.Errorf("Expected:\n")
		t.Errorf("%+v\n", expectedConfig)
		t.Errorf("Actual:\n")
		t.Errorf("%+v\n", actualConfig)

	}

}
