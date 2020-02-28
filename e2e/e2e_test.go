package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/argoproj/argo-events/common"
	gw_v1alpha1 "github.com/argoproj/argo-events/pkg/apis/gateway/v1alpha1"
	sensorv1alpha1 "github.com/argoproj/argo-events/pkg/apis/sensor/v1alpha1"
	sn_v1alpha1 "github.com/argoproj/argo-events/pkg/apis/sensor/v1alpha1"
	esv1alpha1 "github.com/argoproj/argo-events/pkg/client/eventsources/clientset/versioned/typed/eventsources/v1alpha1"
	gwv1alpha1 "github.com/argoproj/argo-events/pkg/client/gateway/clientset/versioned/typed/gateway/v1alpha1"
	snv1alpha1 "github.com/argoproj/argo-events/pkg/client/sensor/clientset/versioned/typed/sensor/v1alpha1"
	wf_v1alpha1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/google/go-github/github"
	"github.com/mitchellh/go-homedir"
	ocv1alpha1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// NOTE: custom resources must be manually added here
func init() {
	if err := wf_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	if err := sn_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	if err := gw_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	if err := gw_v1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	if err := ocv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
}

var (
	registry = runtime.NewEquivalentResourceRegistry()
)

const namespace = "argo-events"

// GitArtifact contains information about an minio stored in git
type GitArtifact struct {
	// Path to file that contains trigger resource definition
	FilePath string `json:"filePath"`
	// Ref to use to pull trigger resource. Will result in a shallow clone and
	// fetch.
	// +optional
	Ref string `json:"ref,omitempty"`
}

type E2ETestSuite struct {
	suite.Suite
	K8sClient         kubernetes.Interface
	GatewayClient     gwv1alpha1.ArgoprojV1alpha1Interface
	SensorClient      snv1alpha1.ArgoprojV1alpha1Interface
	GithubClient      *github.Client
	EventSourceClient esv1alpha1.ArgoprojV1alpha1Interface
	DynamicClient     dynamic.Interface
	Branch            string
	Logger            *logrus.Logger
	ProjDir           string
	DomainName        string
}

func decodeAndUnstructure(b []byte, gvr *metav1.GroupVersionResource) (*unstructured.Unstructured, error) {
	gvk := registry.KindFor(schema.GroupVersionResource{
		Group:    gvr.Group,
		Version:  gvr.Version,
		Resource: gvr.Resource,
	}, "")

	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(b, &gvk, nil)
	if err != nil {
		return nil, err
	}

	uObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: uObj}, nil
}

func encodeWorkflowFromUnstructured(obj *unstructured.Unstructured) (*wf_v1alpha1.Workflow, error) {
	body, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}
	var wf *wf_v1alpha1.Workflow
	if err := json.Unmarshal(body, &wf); err != nil {
		return nil, err
	}
	return wf, nil
}

func wfGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "workflows",
	}
}

func routeGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ocv1alpha1.GroupName,
		Version:  ocv1alpha1.GroupVersion.Version,
		Resource: "routes",
	}
}

func (test *E2ETestSuite) SetupTest() {
	// kubernetes configuration
	kubeConfig, _ := os.LookupEnv(common.EnvVarKubeConfig)
	restConfig, err := common.GetClientConfig(kubeConfig)
	if err != nil {
		panic(err)
	}

	domainName, ok := os.LookupEnv("DOMAIN_NAME")
	if !ok {
		panic("no domain name")
	}

	test.K8sClient = kubernetes.NewForConfigOrDie(restConfig)
	test.GatewayClient = gwv1alpha1.NewForConfigOrDie(restConfig)
	test.SensorClient = snv1alpha1.NewForConfigOrDie(restConfig)
	test.EventSourceClient = esv1alpha1.NewForConfigOrDie(restConfig)
	test.DynamicClient = dynamic.NewForConfigOrDie(restConfig)
	test.GithubClient = github.NewClient(nil)
	test.DomainName = domainName
	test.Branch = "master"
	dir, _ := homedir.Dir()
	test.ProjDir = dir + "/go/src/github.com/argoproj/argo-events"
	test.Logger = common.NewArgoEventsLogger()
}

func (test *E2ETestSuite) WaitForGatewayReadyState(name string) error {
	fmt.Println("waiting for gateway to get ready...")

	sel, err := FieldSelector(map[string]string{
		"metadata.name": name,
	})
	if err != nil {
		return err
	}

	watch, err := test.GatewayClient.Gateways(namespace).Watch(metav1.ListOptions{
		FieldSelector: sel.String(),
	})
	if err != nil {
		return err
	}

	for resource := range watch.ResultChan() {
		gateway := resource.Object.(*gw_v1alpha1.Gateway)
		if gateway.Status.Phase == gw_v1alpha1.NodePhaseRunning {
			break
		}
	}

	return nil
}

func (test *E2ETestSuite) WaitForSensorReadyState(name string) error {
	fmt.Println("waiting for sensor to get ready...")

	sel, err := FieldSelector(map[string]string{
		"metadata.name": name,
	})
	if err != nil {
		return err
	}

	watch, err := test.SensorClient.Sensors(namespace).Watch(metav1.ListOptions{
		FieldSelector: sel.String(),
	})
	if err != nil {
		return err
	}

	for resource := range watch.ResultChan() {
		sensor := resource.Object.(*sensorv1alpha1.Sensor)
		if sensor.Status.Phase == sensorv1alpha1.NodePhaseActive {
			break
		}
	}

	return nil
}

func (test *E2ETestSuite) DeleteAllResources(t *testing.T) {
	err := test.EventSourceClient.EventSources(namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
	require.Nil(t, err)

	err = test.GatewayClient.Gateways(namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
	require.Nil(t, err)

	err = test.SensorClient.Sensors(namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
	require.Nil(t, err)
}

func (test *E2ETestSuite) TestAll() {
	test.CreateControllers(test.T())
	err := test.WaitForControllers()
	require.Nil(test.T(), err)

	test.WebhookSetupTest(test.T())

	test.WebhookResourceTest(test.T(), "webhook-gateway", "webhook-sensor")

	test.DeleteAllResources(test.T())

	err = test.DeleteControllers()
	require.Nil(test.T(), err)
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
