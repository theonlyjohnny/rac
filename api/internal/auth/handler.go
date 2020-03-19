package auth

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/docker/distribution/registry/auth"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
)

var (
	privateKey *rsa.PrivateKey
	kid        string
)

type claims struct {
	Access []*token.ResourceActions `json:"access"`
	jwt.StandardClaims
}

func init() {
	keyPath := "/var/jwt.key"
	privateKeyContents, err := ioutil.ReadFile(keyPath)
	if err != nil {
		panic(fmt.Errorf("Unable to read private key located at %s for signing of JWT -- %s", keyPath, err.Error()))
	}
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyContents)
	if err != nil {
		panic(fmt.Errorf("Unable to parse contents located at %s as RSA Private privateKey -- %s", keyPath, err.Error()))
	}

	publicKey, err := libtrust.FromCryptoPublicKey(&privateKey.PublicKey)

	if err != nil {
		panic(fmt.Errorf("Unable to parse public key from private key -- %s", err.Error()))
	}

	kid = publicKey.KeyID()

}

func Handler(w http.ResponseWriter, req *http.Request) {

	query := req.URL.Query()

	parsed := resolveScopeSpecifiers(query["scope"])

	// fmt.Println(parsed)

	// dockerAccessList := make([]dockerAccess, len(parsed))
	// for i, e := range parsed {
	// dockerAccessList[i] = dockerAccess{
	// Type:    e.Type,
	// Name:    e.Name,
	// Actions: []string{e.Action},
	// }
	// }

	rn := time.Now()

	// []dockerAccess{
	// {
	// Type:    splitScope[0],
	// Name:    splitScope[1],
	// Actions: strings.Split(splitScope[2], ","),
	// },
	// },

	// Make a set of access entries to put in the token's claimset.
	resourceActionSets := make(map[auth.Resource]map[string]struct{}, len(parsed))
	for _, access := range parsed {
		actionSet, exists := resourceActionSets[access.Resource]
		if !exists {
			actionSet = map[string]struct{}{}
			resourceActionSets[access.Resource] = actionSet
		}
		actionSet[access.Action] = struct{}{}
	}

	accessEntries := make([]*token.ResourceActions, 0, len(resourceActionSets))
	for resource, actionSet := range resourceActionSets {
		actions := make([]string, 0, len(actionSet))
		for action := range actionSet {
			actions = append(actions, action)
		}

		accessEntries = append(accessEntries, &token.ResourceActions{
			Type:    resource.Type,
			Class:   resource.Class,
			Name:    resource.Name,
			Actions: actions,
		})
	}

	// Create the Claims
	claims := &claims{
		accessEntries,
		jwt.StandardClaims{
			Audience:  "rac.registry",
			ExpiresAt: rn.Add(1 * time.Hour).Unix(),
			Id:        strconv.FormatInt(rn.UnixNano(), 10),
			IssuedAt:  rn.Unix(),
			Issuer:    "rac.api", //fqdn
			NotBefore: rn.Unix(),
			Subject:   "", //authed user id
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	ss, err := token.SignedString(privateKey)

	if err != nil {
		fmt.Printf("ERR unable to sign JWT: %s \n", err.Error())
	}

	resp := map[string]string{
		"token": ss,
	}

	fmt.Println(ss)

	bytes, err := json.Marshal(resp)
	if err != nil {
		fmt.Printf("ERR: %s \n", err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func parseScope(scope string) (out [3]string) {
	split := strings.Split(scope, ":")
	fmt.Printf("parse :%s into %s \n", scope, split)
	out[0] = split[0]
	out[1] = split[1]
	out[2] = split[2]
	return
}
