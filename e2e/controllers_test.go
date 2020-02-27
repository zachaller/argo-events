package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"testing"

	"github.com/ghodss/yaml"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func (test *E2ETestSuite) CreateGatewayControllerConfigmap() (*corev1.ConfigMap, error) {
	content, err := ioutil.ReadFile(test.ProjDir + "/hack/k8s/manifests/gateway-controller-configmap.yaml")
	if err != nil {
		return nil, err
	}

	var gwCm *corev1.ConfigMap
	if err := yaml.Unmarshal(content, &gwCm); err != nil {
		return nil, err
	}

	return test.K8sClient.CoreV1().ConfigMaps(namespace).Create(gwCm)
}

func (test *E2ETestSuite) CreateSensorControllerConfigmap() (*corev1.ConfigMap, error) {
	content, err := ioutil.ReadFile(test.ProjDir + "/hack/k8s/manifests/sensor-controller-configmap.yaml")
	if err != nil {
		return nil, err
	}

	var snCm *corev1.ConfigMap
	if err := yaml.Unmarshal(content, &snCm); err != nil {
		return nil, err
	}

	return test.K8sClient.CoreV1().ConfigMaps(namespace).Create(snCm)
}

func (test *E2ETestSuite) CreateGatewayController() (*appv1.Deployment, error) {
	content, err := ioutil.ReadFile(test.ProjDir + "/hack/k8s/manifests/gateway-controller-deployment.yaml")
	if err != nil {
		return nil, err
	}

	var gwCtrl *appv1.Deployment
	if err := yaml.Unmarshal(content, &gwCtrl); err != nil {
		return nil, err
	}

	return test.K8sClient.AppsV1().Deployments(namespace).Create(gwCtrl)
}

func (test *E2ETestSuite) CreateSensorController() (*appv1.Deployment, error) {
	content, err := ioutil.ReadFile(test.ProjDir + "/hack/k8s/manifests/sensor-controller-deployment.yaml")
	if err != nil {
		return nil, err
	}

	var snCtrl *appv1.Deployment
	if err := yaml.Unmarshal(content, &snCtrl); err != nil {
		return nil, err
	}

	return test.K8sClient.AppsV1().Deployments(namespace).Create(snCtrl)
}

func (test *E2ETestSuite) CreateControllers(t *testing.T) {
	gwCm, err := test.CreateGatewayControllerConfigmap()
	fmt.Printf("%+v\n", err)
	require.Nil(t, err)
	require.NotEmpty(t, gwCm)

	snCm, err := test.CreateSensorControllerConfigmap()
	require.Nil(t, err)
	require.NotEmpty(t, snCm)

	gwctrl, err := test.CreateGatewayController()
	require.Nil(t, err)
	require.NotNil(t, gwctrl)

	snctrl, err := test.CreateSensorController()
	require.Nil(t, err)
	require.NotNil(t, snctrl)
}

// LabelReq returns label requirements
func LabelReq(key, value string) (*labels.Requirement, error) {
	req, err := labels.NewRequirement(key, selection.Equals, []string{value})
	if err != nil {
		return nil, err
	}
	return req, nil
}

// LabelSelector returns label selector for resource filtering
func LabelSelector(resourceLabels map[string]string) (labels.Selector, error) {
	var labelRequirements []labels.Requirement
	for key, value := range resourceLabels {
		req, err := LabelReq(key, value)
		if err != nil {
			return nil, err
		}
		labelRequirements = append(labelRequirements, *req)
	}
	return labels.NewSelector().Add(labelRequirements...), nil
}

// FieldSelector returns field selector for resource filtering
func FieldSelector(fieldSelectors map[string]string) (fields.Selector, error) {
	var selectors []fields.Selector
	for key, value := range fieldSelectors {
		selector, err := fields.ParseSelector(fmt.Sprintf("%s=%s", key, value))
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, selector)
	}
	return fields.AndSelectors(selectors...), nil
}

func (test *E2ETestSuite) WaitForControllers() error {
	fmt.Println("waiting for deployment to scale...")

	sel, err := FieldSelector(map[string]string{
		"metadata.name": "gateway-controller",
	})
	if err != nil {
		return err
	}

	watch, err := test.K8sClient.AppsV1().Deployments(namespace).Watch(metav1.ListOptions{
		FieldSelector: sel.String(),
	})
	if err != nil {
		return err
	}

	for resource := range watch.ResultChan() {
		dep := resource.Object.(*appv1.Deployment)
		fmt.Printf("%s %d replica update \n", dep.Name, dep.Status.ReadyReplicas)
		if dep.Status.ReadyReplicas == 1 {
			break
		}
	}

	sel, err = FieldSelector(map[string]string{
		"metadata.name": "sensor-controller",
	})
	if err != nil {
		return err
	}

	watch, err = test.K8sClient.AppsV1().Deployments(namespace).Watch(metav1.ListOptions{
		FieldSelector: sel.String(),
	})
	if err != nil {
		return err
	}

	for resource := range watch.ResultChan() {
		dep := resource.Object.(*appv1.Deployment)
		fmt.Printf("%s %d replica update \n", dep.Name, dep.Status.ReadyReplicas)
		if dep.Status.ReadyReplicas == 1 {
			break
		}
	}

	return nil
}

func (test *E2ETestSuite) DeleteControllers() error {
	if err := test.K8sClient.CoreV1().ConfigMaps(namespace).Delete("gateway-controller-configmap", &metav1.DeleteOptions{}); err != nil {
		return err
	}
	if err := test.K8sClient.CoreV1().ConfigMaps(namespace).Delete("sensor-controller-configmap", &metav1.DeleteOptions{}); err != nil {
		return err
	}
	if err := test.K8sClient.AppsV1().Deployments(namespace).Delete("gateway-controller", &metav1.DeleteOptions{}); err != nil {
		return err
	}
	if err := test.K8sClient.AppsV1().Deployments(namespace).Delete("sensor-controller", &metav1.DeleteOptions{}); err != nil {
		return err
	}
	return nil
}
