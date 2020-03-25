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
		return nil, fmt.Errorf("Failed to load ~/.kube/config.yaml: %s", err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Failed to create k8s clientset from config: %s", err.Error())
	}

	deployments := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	tknMgr, err := token.NewTokenManager("/var/jwt.key")
	if err != nil {
		return nil, fmt.Errorf("Failed to create token manager -- %s", err.Error())
	}

	authenticator := auth.NewAuthenticator(dao)

	r := gin.Default()
	//later = run later
	r.Use(
		func(c *gin.Context) {
			q := c.Request.URL.Query()
			c.Request.ParseForm()
			fmt.Printf("path :%s | query: %s | form: %s | headers: %s \n\n", c.Request.URL, q, c.Request.Form, c.Request.Header)
		},
	)

	authMiddleware := func(c *gin.Context) {
		res := auth.ProvidedCredentials{}
		user, password, haveBasicAuth := c.Request.BasicAuth()
		if haveBasicAuth {
			res.Username = user
			res.Password = password
		} else if c.Request.Method == "POST" {
			user, password = c.Request.FormValue("username"), c.Request.FormValue("password")
			if user != "" && password != "" {
				res.Username, res.Password = user, password
			}
			if refresh_token := c.Request.FormValue("refresh_token"); refresh_token != "" {
				res.RefreshToken = refresh_token
			}
		}

		empty := auth.ProvidedCredentials{}

		if res != empty {
			c.Set(authedUserCtxKey, res)
		}
	}

	api := &API{
		dao:   dao,
		token: tknMgr,
		auth:  authenticator,

		clientset:   clientset,
		deployments: deployments,

		Router: r,
	}

	r.POST("/notification", api.postNotification)
	r.POST("/claim", api.postClaim)

	auth := r.Group("/auth")
	{
		auth.Use(authMiddleware)
		auth.GET("/", api.doAuth)
		auth.POST("/", api.doAuth)
	}
	return api, nil
}
