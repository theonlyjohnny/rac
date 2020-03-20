package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/theonlyjohnny/rac/api/internal/auth"
	"github.com/theonlyjohnny/rac/api/internal/storage"
	"github.com/theonlyjohnny/rac/api/internal/token"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	typesv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
)

type API struct {
	dao   storage.DAO
	token token.TokenManager
	auth  auth.Authenticator

	clientset   *kubernetes.Clientset
	deployments typesv1.DeploymentInterface

	Router *gin.Engine
}

func NewAPI() (*API, error) {

	dao := storage.NewDAO()

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config.yaml"))
	if err != nil {
		return nil, fmt.Errorf("Failed to load ~/.kube/config.yaml -- %s", err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to create k8s clientset from config -- %s", err.Error())
	}

	deployments := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	tknMgr, err := token.NewTokenManager("/var/jwt.key")
	if err != nil {
		return nil, fmt.Errorf("Failed to create token manager -- %s", err.Error())
	}

	authenticator := auth.NewAuthenticator(dao)

	r := gin.Default()
	api := &API{
		dao:   dao,
		token: tknMgr,
		auth:  authenticator,

		clientset:   clientset,
		deployments: deployments,

		Router: r,
	}

	r.POST("/notification", api.postNotification)
	r.POST("claim", api.postClaim)
	r.GET("/auth", api.getAuth)
	return api, nil
}
