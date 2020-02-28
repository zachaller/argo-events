package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	eventSourceV1Alpha1 "github.com/argoproj/argo-events/pkg/apis/eventsources/v1alpha1"
	gw_v1alpha1 "github.com/argoproj/argo-events/pkg/apis/gateway/v1alpha1"
	sensorv1alpha1 "github.com/argoproj/argo-events/pkg/apis/sensor/v1alpha1"
	"github.com/ghodss/yaml"
	ocv1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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

	err = test.WaitForGatewayReadyState(gw.Name)
	require.Nil(t, err)

	err = test.WaitForSensorReadyState(sn.Name)
	require.Nil(t, err)
}

func (test *E2ETestSuite) WebhookResourceTest(t *testing.T, gatewayName, sensorName string) {
	gateway, err := test.GatewayClient.Gateways(namespace).Get(gatewayName, metav1.GetOptions{})
	require.Nil(t, err)

	require.Equal(t, gw_v1alpha1.NodePhaseRunning, gateway.Status.Phase)

	sensor, err := test.SensorClient.Sensors(namespace).Get(sensorName, metav1.GetOptions{})
	require.Nil(t, err)

	require.Equal(t, sensorv1alpha1.NodePhaseActive, sensor.Status.Phase)

	gwservice, err := test.K8sClient.CoreV1().Services(namespace).Get(fmt.Sprintf("%s-svc", gatewayName), metav1.GetOptions{})
	require.Nil(t, err)

	route := &ocv1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name: gwservice.Name,
		},
		Spec: ocv1.RouteSpec{
			Host: "webhook-gateway-svc-argo-events.",
			To: ocv1.RouteTargetReference{
				Kind: "Service",
				Name: gwservice.Name,
			},
			Port: &ocv1.RoutePort{TargetPort: intstr.FromString("example")},
		},
	}

	routeBody, err := yaml.Marshal(route)
	require.Nil(t, err)
	require.NotNil(t, routeBody)

	gvr := routeGVR()
	mgvr := &metav1.GroupVersionResource{
		Group:    gvr.Group,
		Version:  gvr.Version,
		Resource: gvr.Resource,
	}

	routeUObj, err := decodeAndUnstructure(routeBody, mgvr)
	require.Nil(t, err)

	routeClient := test.DynamicClient.Resource(routeGVR()).Namespace(namespace)

	routeNewObj, err := routeClient.Create(routeUObj, metav1.CreateOptions{})
	require.Nil(t, err)
	require.NotNil(t, routeNewObj)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	reader := bytes.NewReader([]byte(`{"message":"this is my first webhook"}`))

	response, err := client.Post("webhook-gateway-svc-argo-events./example", "application/json", reader)
	require.Nil(t, err)
	body, err := ioutil.ReadAll(response.Body)
	require.Nil(t, err)
	require.Equal(t, "success", string(body))

	wfClient := test.DynamicClient.Resource(wfGVR()).Namespace(namespace)
	wfs, err := wfClient.List(metav1.ListOptions{})
	require.Nil(t, err)
	require.Equal(t, 1, len(wfs.Items))

	wfUObj := wfs.Items[0]
	wf, err := encodeWorkflowFromUnstructured(&wfUObj)
	require.Nil(t, err)

	require.Equal(t, true, strings.Contains(wf.Name, "webhook-"))
	require.NotEqual(t, "hello world", wf.Spec.Arguments.Parameters[0].Value)

	err = routeClient.Delete(gwservice.Name, &metav1.DeleteOptions{})
	require.Nil(t, err)
}
