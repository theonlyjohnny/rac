package token

import (
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/docker/distribution/registry/auth"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
)

type TokenManager interface {
	CreateTokenForAcess(access []auth.Access) (string, error)
}

type tokenManagerImpl struct {
	privateKey *rsa.PrivateKey
	kid        string

	issuer   string
	audience string
}

type fullClaims struct {
	Access []*token.ResourceActions `json:"access"`
	jwt.StandardClaims
}

func NewTokenManager(keyPath string) (TokenManager, error) {

	privateKeyContents, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read private key located at %s for signing of JWT -- %s", keyPath, err.Error())
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyContents)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse contents located at %s as RSA Private privateKey -- %s", keyPath, err.Error())
	}

	publicKey, err := libtrust.FromCryptoPublicKey(&privateKey.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("Unable to parse public key from private key -- %s", err.Error())
	}

	kid := publicKey.KeyID()
	return &tokenManagerImpl{
		privateKey: privateKey,
		kid:        kid,
		issuer:     "rac.api", //fqdn
		audience:   "rac.registry",
	}, nil
}

func (tm *tokenManagerImpl) CreateTokenForAcess(access []auth.Access) (string, error) {

	accessEntries := accessToAccessEntries(access)

	rn := time.Now()
	claims := &fullClaims{
		accessEntries,
		jwt.StandardClaims{
			Audience:  tm.audience,
			ExpiresAt: rn.Add(1 * time.Hour).Unix(),
			Id:        strconv.FormatInt(rn.UnixNano(), 10),
			IssuedAt:  rn.Unix(),
			Issuer:    tm.issuer,
			NotBefore: rn.Unix(),
			Subject:   "", //authed user id
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = tm.kid
	ss, err := token.SignedString(tm.privateKey)

	if err != nil {
		return "", fmt.Errorf("unable to sign JWT: %s", err.Error())
	}

	return ss, nil
}

func accessToAccessEntries(input []auth.Access) []*token.ResourceActions {

	// Make a set of access entries to put in the token's claimset.
	resourceActionSets := make(map[auth.Resource]map[string]struct{}, len(input))
	// {
	//	 resource: {
	//		 [action]: struct{}
	//	 }
	// }
	for _, access := range input {
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

	return accessEntries
}
