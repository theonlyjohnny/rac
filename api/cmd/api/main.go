package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/theonlyjohnny/rac/api/internal/auth"
	"github.com/theonlyjohnny/rac/api/internal/notification"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config.yaml"))
	if err != nil {
		panic(err.Error())
	}

	// fmt.Printf("%#v \n", config)

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	deployments := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	http.HandleFunc("/notification", notification.Handler(deployments))
	http.HandleFunc("/auth", auth.Handler)

	fmt.Println("running")

	if err := http.ListenAndServe(":8090", nil); err != nil {
		panic(err)
	}
}
