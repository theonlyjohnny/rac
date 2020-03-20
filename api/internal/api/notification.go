package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/theonlyjohnny/rac/api/internal/storage"

	"github.com/docker/distribution/notifications"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (a *API) postNotification(c *gin.Context) {
	var body notifications.Envelope
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, e := range body.Events {
		if e.Action == "push" {
			for _, d := range e.Target.References {
				if d.MediaType == "application/vnd.docker.container.image.v1+json" {
					fmt.Printf("Pushed tag %s of %s from %s \n", e.Target.Tag, e.Target.Repository, e.Request.Addr)
					repo := &storage.Repo{Name: e.Target.Repository}
					if err := a.dao.SaveRepo(repo); err != nil {
						fmt.Printf("Failed to save repo %v -- %s\n", repo, err.Error())
						return
					}
					if err := a.triggerUpdate(e.Target.Repository, e.Target.Tag); err != nil {
						fmt.Printf("Failed to trigger update for repo %v -- %s\n", repo, err.Error())
						return
					}
				}
			}
		}
	}
}

func (a *API) triggerUpdate(repo, tag string) error {
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

	_, err := a.deployments.Create(deployment)
	if err != nil {
		return fmt.Errorf("Failed to create deployment: %s", err.Error())
	}

	fmt.Printf("Creating deployment of %s:%s \n", repo, tag)

	return nil
}

func int32Ptr(i int32) *int32 { return &i }
