package notification

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apptypesv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func TriggerUpdate(deployments apptypesv1.DeploymentInterface, repo, tag string) error {
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
							//TODO read from image
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

	_, err := deployments.Create(deployment)
	if err != nil {
		return fmt.Errorf("Failed to create deployment: %s", err.Error())
	}

	fmt.Printf("Creating deployment of %s:%s \n", repo, tag)

	return nil

	// fmt.Printf("res: %v\n", res)
}

func int32Ptr(i int32) *int32 { return &i }
