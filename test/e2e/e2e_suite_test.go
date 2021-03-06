package e2e_test

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	logs "github.com/appscode/go/log/golog"
	api "github.com/appscode/stash/apis/stash"
	"github.com/appscode/stash/client/scheme"
	_ "github.com/appscode/stash/client/scheme"
	cs "github.com/appscode/stash/client/typed/stash/v1alpha1"
	"github.com/appscode/stash/pkg/controller"
	"github.com/appscode/stash/test/e2e/framework"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	crd_cs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	TIMEOUT             = 20 * time.Minute
	TestSidecarImageTag = "canary"
)

var (
	ctrl             *controller.StashController
	root             *framework.Framework
	createInitConfig bool
)

func init() {
	// ./hack/make.py test e2e -init-config=true -v=3
	flag.BoolVar(&createInitConfig, "init-config", false, "create initializer config")
}

func TestE2e(t *testing.T) {
	logs.InitLogs()
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(TIMEOUT)
	junitReporter := reporters.NewJUnitReporter("junit.xml")
	RunSpecsWithDefaultAndCustomReporters(t, "e2e Suite", []Reporter{junitReporter})
}

var _ = BeforeSuite(func() {
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube/config")
	By("Using kubeconfig from " + kubeconfigPath)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	Expect(err).NotTo(HaveOccurred())

	kubeClient := kubernetes.NewForConfigOrDie(config)
	stashClient := cs.NewForConfigOrDie(config)
	crdClient := crd_cs.NewForConfigOrDie(config)
	scheme.AddToScheme(clientsetscheme.Scheme)

	root = framework.New(kubeClient, stashClient)
	err = root.CreateNamespace()
	Expect(err).NotTo(HaveOccurred())
	By("Using test namespace " + root.Namespace())

	opts := controller.Options{
		SidecarImageTag: TestSidecarImageTag,
		ResyncPeriod:    5 * time.Minute,
	}
	ctrl = controller.New(kubeClient, crdClient, stashClient, opts)
	By("Registering CRD group " + api.GroupName)
	err = ctrl.Setup()
	Expect(err).NotTo(HaveOccurred())
	root.EventuallyCRD("restic." + api.GroupName).Should(Succeed())

	if createInitConfig {
		By("Creating workload initializer")
		root.CreateInitializerConfiguration(root.InitializerForWorkloads())
	}

	// Now let's start the controller
	// stop := make(chan struct{})
	// defer close(stop)
	go ctrl.Run(1, nil)
})

var _ = AfterSuite(func() {
	root.DeleteNamespace()
	if createInitConfig {
		root.DeleteInitializerConfiguration(root.InitializerForWorkloads().ObjectMeta)
	}
})
