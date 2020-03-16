package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/docker/distribution/notifications"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apptypesv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	clientset   *kubernetes.Clientset
	deployments apptypesv1.DeploymentInterface
)

func mainFunc(w http.ResponseWriter, req *http.Request) {
	var res notifications.Envelope

	err := json.NewDecoder(req.Body).Decode(&res)
	if err != nil {
		panic(err)
	}

	for _, e := range res.Events {
		if e.Action == "push" {
			for _, d := range e.Target.References {
				if d.MediaType == "application/vnd.docker.container.image.v1+json" {
					fmt.Printf("Pushed tag %s of %s from %s \n", e.Target.Tag, e.Target.Repository, e.Request.Addr)
					triggerUpdate(e.Target.Repository, e.Target.Tag)
				}
			}
		}
	}
}

func triggerUpdate(repo, tag string) {
	fmt.Println("TODO")
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: repo,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": repo,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": repo,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  repo,
							Image: fmt.Sprintf("rac.registry:5000/%s:%s", repo, tag),
							// Ports: []apiv1.ContainerPort{
							// {
							// Name:          "http",
							// Protocol:      apiv1.ProtocolTCP,
							// ContainerPort: 80,
							// },
							// },
						},
					},
				},
			},
		},
	}
	fmt.Printf("Creating %v \n", deployment)
	res, err := deployments.Create(deployment)
	if err != nil {
		fmt.Printf("Failed to create deployment: %s \n", err.Error())
		return
	}

	fmt.Printf("res: %v\n", res)
}

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config.yaml"))
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("%#v \n", config)

	// create the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// fmt.Printf("rest: %#v \nclient: %#v \n", rCli, rCli.Client)

	deployments = clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	http.HandleFunc("/", mainFunc)

	fmt.Println("running")

	if err := http.ListenAndServe(":8090", nil); err != nil {
		panic(err)
	}
}

func int32Ptr(i int32) *int32 { return &i }
