package main

import (
	"context"
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/go-github/github"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func (test *E2ETestSuite) CreateGatewayControllerConfigmap() (*corev1.ConfigMap, error) {
	content, _, _, err := test.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "hack/k8s/manifests/gateway-controller-configmap.yaml", &github.RepositoryContentGetOptions{
		Ref: test.Branch,
	})
	if err != nil {
		return nil, err
	}

	gwCmBody, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return nil, err
	}

	var gwCm *corev1.ConfigMap
	if err := yaml.Unmarshal(gwCmBody, &gwCm); err != nil {
		return nil, err
	}

	return test.K8sClient.CoreV1().ConfigMaps(namespace).Create(gwCm)
}

func (test *E2ETestSuite) CreateSensorControllerConfigmap() (*corev1.ConfigMap, error) {
	content, _, _, err := test.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "hack/k8s/manifests/sensor-controller-configmap.yaml", &github.RepositoryContentGetOptions{
		Ref: test.Branch,
	})
	if err != nil {
		return nil, err
	}

	snCmBody, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return nil, err
	}

	var snCm *corev1.ConfigMap
	if err := yaml.Unmarshal(snCmBody, &snCm); err != nil {
		return nil, err
	}

	return test.K8sClient.CoreV1().ConfigMaps(namespace).Create(snCm)
}

func (test *E2ETestSuite) CreateGatewayController() (*appv1.Deployment, error) {
	content, _, _, err := test.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "hack/k8s/manifests/gateway-controller-deployment.yaml", &github.RepositoryContentGetOptions{
		Ref: test.Branch,
	})
	if err != nil {
		return nil, err
	}

	gwCtrlBody, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return nil, err
	}

	var gwCtrl *appv1.Deployment
	if err := yaml.Unmarshal(gwCtrlBody, &gwCtrl); err != nil {
		return nil, err
	}

	return test.K8sClient.AppsV1().Deployments(namespace).Create(gwCtrl)
}

func (test *E2ETestSuite) CreateSensorController() (*appv1.Deployment, error) {
	content, _, _, err := test.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "hack/k8s/manifests/sensor-controller-deployment.yaml", &github.RepositoryContentGetOptions{
		Ref: test.Branch,
	})
	if err != nil {
		return nil, err
	}

	snCtrlBody, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return nil, err
	}

	var snCtrl *appv1.Deployment
	if err := yaml.Unmarshal(snCtrlBody, &snCtrl); err != nil {
		return nil, err
	}

	return test.K8sClient.AppsV1().Deployments(namespace).Create(snCtrl)
}

func (test *E2ETestSuite) ControllersTest(t *testing.T) {
	gwCm, err := test.CreateGatewayControllerConfigmap()
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
