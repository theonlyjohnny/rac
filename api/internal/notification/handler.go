package notification

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/docker/distribution/notifications"
	apptypesv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
)

func Handler(deployments apptypesv1.DeploymentInterface) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
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
						triggerUpdate(deployments, e.Target.Repository, e.Target.Tag)
					}
				}
			}
		}
	}
}
