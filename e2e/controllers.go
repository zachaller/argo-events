package main

import (
	"context"
	"encoding/base64"
	"github.com/ghodss/yaml"
	"github.com/google/go-github/github"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func (suite *E2ETestSuite) CreateControllerConfigmaps() error {

	content, _, _, err := suite.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "hack/k8s/manifests/gateway-controller-configmap.yaml", &github.RepositoryContentGetOptions{
		Ref: suite.Branch,
	})
	if err != nil {
		return err
	}

	gwCmBody, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return err
	}

	var gwCm *corev1.ConfigMap
	if err := yaml.Unmarshal(gwCmBody, &gwCm); err != nil {
		return err
	}

	if _, err = suite.K8sClient.CoreV1().ConfigMaps(namespace).Create(gwCm); err != nil {
		return err
	}

	content, _, _, err = suite.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "hack/k8s/manifests/sensor-controller-deployment.yaml", &github.RepositoryContentGetOptions{
		Ref: suite.Branch,
	})
	if err != nil {
		return err
	}

	snCtrlBody, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return err
	}

	var snCtrl *appv1.Deployment
	if err := yaml.Unmarshal(snCtrlBody, &snCtrl); err != nil {
		return err
	}

	return nil
}

func (suite *E2ETestSuite) CreateControllers() error {
	content, _, _, err := suite.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "hack/k8s/manifests/gateway-controller-deployment.yaml", &github.RepositoryContentGetOptions{
		Ref: suite.Branch,
	})
	if err != nil {
		return err
	}

	gwCtrlBody, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return err
	}

	var gwCtrl *appv1.Deployment
	if err := yaml.Unmarshal(gwCtrlBody, &gwCtrl); err != nil {
		return err
	}

	if _, err = suite.K8sClient.AppsV1().Deployments(namespace).Create(gwCtrl); err != nil {
		return err
	}

	content, _, _, err = suite.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "hack/k8s/manifests/sensor-controller-deployment.yaml", &github.RepositoryContentGetOptions{
		Ref: suite.Branch,
	})
	if err != nil {
		return err
	}

	snCtrlBody, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return err
	}

	var snCtrl *appv1.Deployment
	if err := yaml.Unmarshal(snCtrlBody, &snCtrl); err != nil {
		return err
	}

	return nil
}
