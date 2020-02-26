package main

import (
	"context"
	"encoding/base64"
	eventSourceV1Alpha1 "github.com/argoproj/argo-events/pkg/apis/eventsources/v1alpha1"
	gw_v1alpha1 "github.com/argoproj/argo-events/pkg/apis/gateway/v1alpha1"
	sensorv1alpha1 "github.com/argoproj/argo-events/pkg/apis/sensor/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/google/go-github/github"
)

func (suite *E2ETestSuite) CreateWebhookEventSource() (*eventSourceV1Alpha1.EventSource, error) {
	content, _, _, err := suite.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "examples/event-sources/webhook.yaml", &github.RepositoryContentGetOptions{
		Ref: suite.Branch,
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

	return suite.EventSourceClient.EventSources(namespace).Create(es)
}

func (suite *E2ETestSuite) CreateWebhookGateway() (*gw_v1alpha1.Gateway, error) {
	content, _, _, err := suite.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "examples/gateways/webhook.yaml", &github.RepositoryContentGetOptions{
		Ref: suite.Branch,
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
	return suite.GatewayClient.Gateways(namespace).Create(gw)
}

func (suite *E2ETestSuite) CreateWebhookSensor() (*sensorv1alpha1.Sensor, error) {
	content, _, _, err := suite.GithubClient.Repositories.GetContents(context.Background(), "argoproj", "argo-events", "examples/sensors/webhook.yaml", &github.RepositoryContentGetOptions{
		Ref: suite.Branch,
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
	return suite.SensorClient.Sensors(namespace).Create(sensor)
}

func TestWebhook() error {

}
