package keydb

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwk"
	"gitlab.com/daedaluz/gindb"
	"time"
)

type ServerKey struct {
	ID        string    `json:"id" db:"kid"`
	Type      string    `json:"type" db:"type"`
	Algorithm string    `json:"alg" db:"alg"`
	Created   time.Time `json:"created" db:"created"`
	Private   string    `json:"client" db:"private_key"`
	Public    string    `json:"public" db:"public_key"`
}

func (s *ServerKey) GetPrivateJWK() (jwk.Key, error) {
	return jwk.ParseKey([]byte(s.Private))
}

func (s *ServerKey) GetPublicJWK() (jwk.Key, error) {
	return jwk.ParseKey([]byte(s.Public))
}

func (s *ServerKey) GetSigner() (crypto.Signer, error) {
	var res crypto.Signer
	switch s.Algorithm {
	case "EdDSA":
		var key ed25519.PrivateKey
		if err := jwk.ParseRawKey([]byte(s.Private), &key); err != nil {
			return nil, err
		}
		res = key
	case "RS256", "RS384", "RS512":
		var key rsa.PrivateKey
		if err := jwk.ParseRawKey([]byte(s.Private), &key); err != nil {
			return nil, err
		}
		res = &key
	case "ES256", "ES384", "ES512":
		var key ecdsa.PrivateKey
		if err := jwk.ParseRawKey([]byte(s.Private), &key); err != nil {
			return nil, err
		}
		res = &key
	}
	return res, nil
}

type ServerKeyList []*ServerKey

func (l ServerKeyList) Set() (jwk.Set, error) {
	set := jwk.NewSet()
	for _, k := range l {
		j, err := k.GetPublicJWK()
		if err != nil {
			return nil, err
		}
		set.Add(j)
	}
	return set, nil
}

func GetKeys(c *gin.Context) (ServerKeyList, error) {
	res := make([]*ServerKey, 0, 10)
	tx := gindb.GetTX(c)
	err := tx.Select(&res, `call list_server_keys()`)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetKey(c *gin.Context, kid string) (*ServerKey, error) {
	tx := gindb.GetTX(c)
	res := &ServerKey{}
	row := tx.QueryRowx(`call get_server_key(?)`, kid)
	err := row.StructScan(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetFirstWithAlg(c *gin.Context, alg string) (*ServerKey, error) {
	tx := gindb.GetTX(c)
	res := &ServerKey{}
	row := tx.QueryRowx(`call get_server_key_with_alg(?)`, alg)
	err := row.Scan(&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DeleteKey(c *gin.Context, kid string) error {
	tx := gindb.GetTX(c)
	_, err := tx.Exec(`call delete_server_key(?)`, kid)
	if err != nil {
		return err
	}
	return nil
}

func GetAvailableAlgorithms(c *gin.Context) ([]string, error) {
	res := make([]string, 0, 10)
	tx := gindb.GetTX(c)
	err := tx.Select(&res, `call get_available_algorithms()`)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CreateKey(c *gin.Context, pKey, pubKey jwk.Key) error {
	kid := pKey.KeyID()
	if kid == "" {
		hkid, err := pubKey.Thumbprint(crypto.SHA256)
		if err != nil {
			return err
		}
		kid = hex.EncodeToString(hkid[:8])
		if err := pKey.Set(jwk.KeyIDKey, kid); err != nil {
			return err
		}
		if err := pubKey.Set(jwk.KeyIDKey, kid); err != nil {
			return err
		}
		if err := pKey.Set(jwk.KeyUsageKey, "sig"); err != nil {
			return err
		}
		if err := pubKey.Set(jwk.KeyUsageKey, "sig"); err != nil {
			return err
		}
	}
	typ := pKey.KeyType()
	alg := pKey.Algorithm()

	pKeyStr, _ := json.Marshal(pKey)
	pubKeyStr, _ := json.Marshal(pubKey)

	tx := gindb.GetTX(c)
	_, err := tx.Exec(`call create_server_key(?, ?, ?, ?, ?)`, kid, typ, alg, pKeyStr, pubKeyStr)
	if err != nil {
		return err
	}
	return nil
}
