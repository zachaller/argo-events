package main

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	eventSourceV1Alpha1 "github.com/argoproj/argo-events/pkg/apis/eventsources/v1alpha1"
	gw_v1alpha1 "github.com/argoproj/argo-events/pkg/apis/gateway/v1alpha1"
	sensorv1alpha1 "github.com/argoproj/argo-events/pkg/apis/sensor/v1alpha1"
	"github.com/ghodss/yaml"
)

func (test *E2ETestSuite) CreateWebhookEventSource() (*eventSourceV1Alpha1.EventSource, error) {
	content, err := ioutil.ReadFile(test.ProjDir + "/examples/event-sources/webhook.yaml")
	if err != nil {
		return nil, err
	}

	var es *eventSourceV1Alpha1.EventSource
	if err := yaml.Unmarshal(content, &es); err != nil {
		return nil, err
	}

	return test.EventSourceClient.EventSources(namespace).Create(es)
}

func (test *E2ETestSuite) CreateWebhookGateway() (*gw_v1alpha1.Gateway, error) {
	content, err := ioutil.ReadFile(test.ProjDir + "/examples/gateways/webhook.yaml")
	if err != nil {
		return nil, err
	}

	var gw *gw_v1alpha1.Gateway
	if err := yaml.Unmarshal(content, &gw); err != nil {
		return nil, err
	}
	return test.GatewayClient.Gateways(namespace).Create(gw)
}

func (test *E2ETestSuite) CreateWebhookSensor() (*sensorv1alpha1.Sensor, error) {
	content, err := ioutil.ReadFile(test.ProjDir + "/examples/sensors/webhook.yaml")
	if err != nil {
		return nil, err
	}

	var sensor *sensorv1alpha1.Sensor
	if err := yaml.Unmarshal(content, &sensor); err != nil {
		return nil, err
	}
	return test.SensorClient.Sensors(namespace).Create(sensor)
}

func (test *E2ETestSuite) WebhookSetupTest(t *testing.T) {
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

func (test *E2ETestSuite) WebhookResourceTest(t *testing.T, name string) {
	gateway, err := test.GatewayClient.Gateways(namespace).Get(name, metav1.GetOptions{})
	require.Nil(t, err)

	require.Equal(t, gw_v1alpha1.NodePhaseRunning, gateway.Status.Phase)
	require.Equal(t, 1, len(gateway.Status.Nodes))
}
