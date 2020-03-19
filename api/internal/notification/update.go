package notification

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apptypesv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func triggerUpdate(deployments apptypesv1.DeploymentInterface, repo, tag string) {
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

func int32Ptr(i int32) *int32 { return &i }
