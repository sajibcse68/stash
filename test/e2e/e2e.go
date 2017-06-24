package test

import (
	"errors"
	"os/user"
	"path/filepath"
	"time"

	"github.com/appscode/log"
	rcs "github.com/appscode/stash/client/clientset"
	"github.com/appscode/stash/pkg/controller"
	"github.com/appscode/stash/pkg/eventer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	clientset "k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/tamalsaha/go-oneliners"
	"k8s.io/client-go/rest"
	"fmt"
	"net/http"
)

var image = "appscode/stash:latest"

func runController() (*controller.Controller, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(usr.HomeDir, ".kube/config"))
	if err != nil {
		return &controller.Controller{}, err
	}
	kubeClient := clientset.NewForConfigOrDie(config)
	stashClient := rcs.NewForConfigOrDie(config)


	cc := kubeClient.CoreV1().RESTClient().(*rest.RESTClient)
	cc.Client.Transport = NewWrapperRT(cc.Client)


	ctrl := controller.NewController(kubeClient, stashClient, "canary")
	oneliners.FILE()
	if err := ctrl.Setup(); err != nil {
		log.Errorln(err)
	}
	ctrl.Run()
	return ctrl, nil
}


type WrapperRT struct {
	i http.RoundTripper
}

func NewWrapperRT(c *http.Client) http.RoundTripper {
	if c != nil && c.Transport != nil {
		return &WrapperRT{i: c.Transport}
	}
	return &WrapperRT{i: http.DefaultTransport}
}

func (w WrapperRT) RoundTrip(r *http.Request) (*http.Response, error) {
	fmt.Println(r.Method + "|" + r.URL.String())
	return w.i.RoundTrip(r)
}


func checkEventForBackup(ctrl *controller.Controller, objName string) error {
	var err error
	try := 0
	sets := fields.Set{
		"involvedObject.kind":      "Stash",
		"involvedObject.name":      objName,
		"involvedObject.namespace": namespace,
		"type": apiv1.EventTypeNormal,
	}
	fieldSelector := fields.SelectorFromSet(sets)
	for {
		events, err := ctrl.KubeClient.CoreV1().Events(namespace).List(metav1.ListOptions{FieldSelector: fieldSelector.String()})
		if err == nil {
			for _, e := range events.Items {
				if e.Reason == eventer.EventReasonSuccessfulBackup {
					return nil
				}
			}
		}
		if try > 12 {
			return err
		}
		log.Infoln("Waiting for 10 second for events of backup process")
		time.Sleep(time.Second * 10)
		try++
	}
	return errors.New("Stash backup failed.")
	return err
}

func checkContainerAfterBackupDelete(watcher *controller.Controller, name string, _type string) error {
	try := 0
	var err error
	var containers []apiv1.Container
	for {
		log.Infoln("Waiting 20 sec for checking stash-sidecar deletion")
		time.Sleep(time.Second * 20)
		switch _type {
		case controller.ReplicationController:
			rc, err := watcher.KubeClient.CoreV1().ReplicationControllers(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				containers = rc.Spec.Template.Spec.Containers
			}
		case controller.ReplicaSet:
			rs, err := watcher.KubeClient.ExtensionsV1beta1().ReplicaSets(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				containers = rs.Spec.Template.Spec.Containers
			}
		case controller.Deployment:
			deployment, err := watcher.KubeClient.ExtensionsV1beta1().Deployments(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				containers = deployment.Spec.Template.Spec.Containers
			}

		case controller.DaemonSet:
			daemonset, err := watcher.KubeClient.ExtensionsV1beta1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				containers = daemonset.Spec.Template.Spec.Containers
			}
		}
		err = checkContainerDeletion(containers)
		if err == nil {
			break
		}
		try++
		if try > 6 {
			break
		}
	}
	return err
}

func checkContainerDeletion(containers []apiv1.Container) error {
	for _, c := range containers {
		if c.Name == controller.ContainerName {
			return errors.New("ERROR: Stash sidecar not deleted")
		}
	}
	return nil
}
