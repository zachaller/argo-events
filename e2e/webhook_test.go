package main

import (
	"context"
	"encoding/base64"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	eventSourceV1Alpha1 "github.com/argoproj/argo-events/pkg/apis/eventsources/v1alpha1"
	gw_v1alpha1 "github.com/argoproj/argo-events/pkg/apis/gateway/v1alpha1"
	sensorv1alpha1 "github.com/argoproj/argo-events/pkg/apis/sensor/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/google/go-github/github"
)

func (test *E2ETestSuite) CreateWebhookEventSource() (*eventSourceV1Alpha1.EventSource, error) {
	content, _, _, err := test.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "examples/event-sources/webhook.yaml", &github.RepositoryContentGetOptions{
		Ref: test.Branch,
	})
	if err != nil {
		return nil, err
	}

	body, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return nil, err
	}

	var es *eventSourceV1Alpha1.EventSource
	if err := yaml.Unmarshal(body, &es); err != nil {
		return nil, err
	}

	return test.EventSourceClient.EventSources(namespace).Create(es)
}

func (test *E2ETestSuite) CreateWebhookGateway() (*gw_v1alpha1.Gateway, error) {
	content, _, _, err := test.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "examples/gateways/webhook.yaml", &github.RepositoryContentGetOptions{
		Ref: test.Branch,
	})
	if err != nil {
		return nil, err
	}

	body, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return nil, err
	}

	var gw *gw_v1alpha1.Gateway
	if err := yaml.Unmarshal(body, &gw); err != nil {
		return nil, err
	}
	return test.GatewayClient.Gateways(namespace).Create(gw)
}

func (test *E2ETestSuite) CreateWebhookSensor() (*sensorv1alpha1.Sensor, error) {
	content, _, _, err := test.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "examples/sensors/webhook.yaml", &github.RepositoryContentGetOptions{
		Ref: test.Branch,
	})
	if err != nil {
		return nil, err
	}

	body, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		return nil, err
	}

	var sensor *sensorv1alpha1.Sensor
	if err := yaml.Unmarshal(body, &sensor); err != nil {
		return nil, err
	}
	return test.SensorClient.Sensors(namespace).Create(sensor)
}

func (test *E2ETestSuite) WebhookTest(t *testing.T) {
	es, err := test.CreateWebhookEventSource()
	require.Nil(t, err)
	require.NotEmpty(t, es)

	gw, err := test.CreateWebhookGateway()
	require.Nil(t, err)
	require.NotEmpty(t, gw)

	sn, err := test.CreateWebhookSensor()
	require.Nil(t, err)
	require.NotEmpty(t, sn)
}

func (test *E2ETestSuite) DeleteWebhookEventSource(name string) {
	test.EventSourceClient.EventSources(namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
}
