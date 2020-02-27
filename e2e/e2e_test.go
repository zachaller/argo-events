package main

import (
	"github.com/argoproj/argo-events/common"
	esv1alpha1 "github.com/argoproj/argo-events/pkg/client/eventsources/clientset/versioned/typed/eventsources/v1alpha1"
	gwv1alpha1 "github.com/argoproj/argo-events/pkg/client/gateway/clientset/versioned/typed/gateway/v1alpha1"
	snv1alpha1 "github.com/argoproj/argo-events/pkg/client/sensor/clientset/versioned/typed/sensor/v1alpha1"
	wfv1alpha1 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/google/go-github/github"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"testing"
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
	WorkflowClient    wfv1alpha1.ArgoprojV1alpha1Interface
	GithubClient      *github.Client
	EventSourceClient esv1alpha1.ArgoprojV1alpha1Interface
	Branch            string
	Logger            *logrus.Logger
	ProjDir           string
}

func (test *E2ETestSuite) SetupTest() {
	os.Setenv(common.EnvVarKubeConfig, "/Users/vpage/.kube/config")
	os.Setenv("KUBERNETES_SERVICE_HOST", "https://dev-k8s-ewd.dev.blackrock.com")
	os.Setenv("KUBERNETES_SERVICE_PORT", "8443")

	// kubernetes configuration
	kubeConfig, _ := os.LookupEnv(common.EnvVarKubeConfig)
	restConfig, err := common.GetClientConfig(kubeConfig)
	if err != nil {
		panic(err)
	}

	test.K8sClient = kubernetes.NewForConfigOrDie(restConfig)
	test.GatewayClient = gwv1alpha1.NewForConfigOrDie(restConfig)
	test.SensorClient = snv1alpha1.NewForConfigOrDie(restConfig)
	test.WorkflowClient = wfv1alpha1.NewForConfigOrDie(restConfig)
	test.EventSourceClient = esv1alpha1.NewForConfigOrDie(restConfig)
	test.GithubClient = github.NewClient(nil)
	test.Branch = "master"
	dir, _ := homedir.Dir()
	test.ProjDir = dir + "/go/src/github.com/argoproj/argo-events"
	test.Logger = common.NewArgoEventsLogger()
}

func (test *E2ETestSuite) DeleteAll(t *testing.T) {
	err := test.EventSourceClient.EventSources(namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
	require.Nil(t, err)

	err = test.GatewayClient.Gateways(namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
	require.Nil(t, err)

	err = test.SensorClient.Sensors(namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{})
	require.Nil(t, err)
}

func (test *E2ETestSuite) TestAll() {
	test.CreateControllers(test.T())
	werr := test.WaitForControllers()
	require.Nil(test.T(), werr)

	test.WebhookSetupTest(test.T())
	test.DeleteAll(test.T())

	err := test.DeleteControllers()
	require.Nil(test.T(), err)
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
