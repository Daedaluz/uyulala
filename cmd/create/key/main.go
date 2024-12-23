package key

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/daedaluz/gindb"
)

var (
	KeyAlg *string
	db     *sqlx.DB
)

func createRSXKey(length int) {
	privateKey, err := rsa.GenerateKey(rand.Reader, length*8)
	if err != nil {
		slog.Error("Couldn't generate RSA key", "error", err)
		os.Exit(1)
	}
	publicKey := privateKey.Public()
	privateJWK, err := jwk.New(privateKey)
	if err != nil {
		slog.Error("Couldn't generate JWK", "error", err)
		os.Exit(1)
	}
	publicJWK, err := jwk.New(publicKey)
	if err != nil {
		slog.Error("Couldn't generate JWK", "error", err)
		os.Exit(1)
	}
	kidHash, err := publicJWK.Thumbprint(crypto.SHA256)
	if err != nil {
		slog.Error("Couldn't generate JWK KID", "error", err)
		os.Exit(1)
	}
	kid := fmt.Sprintf("%X", kidHash[0:8])
	if err := privateJWK.Set("alg", fmt.Sprintf("RS%d", length)); err != nil {
		slog.Error("Couldn't set JWK alg", "error", err)
		os.Exit(1)
	}
	if err := publicJWK.Set("alg", fmt.Sprintf("RS%d", length)); err != nil {
		slog.Error("Couldn't set JWK alg", "error", err)
		os.Exit(1)
	}
	if err := privateJWK.Set("use", "sig"); err != nil {
		slog.Error("Couldn't set JWK use", "error", err)
		os.Exit(1)
	}
	if err := publicJWK.Set("use", "sig"); err != nil {
		slog.Error("Couldn't set JWK use", "error", err)
		os.Exit(1)
	}
	if err := privateJWK.Set("kid", kid); err != nil {
		slog.Error("Couldn't set JWK kid", "error", err)
		os.Exit(1)
	}
	if err := publicJWK.Set("kid", kid); err != nil {
		slog.Error("Couldn't set JWK kid", "error", err)
		os.Exit(1)
	}

	privateString, _ := json.Marshal(privateJWK)
	publicString, _ := json.Marshal(publicJWK)

	tx := db.MustBegin()
	if _, err := tx.Exec(`call create_server_key(?, ?, ?, ?, ?)`, kid, "RSA", fmt.Sprintf("RS%d", length), privateString, publicString); err != nil {
		slog.Error("Couldn't create server key", "error", err)
		os.Exit(1)
	}
	if err := tx.Commit(); err != nil {
		slog.Error("Couldn't commit transaction", "error", err)
		os.Exit(1)
	}
	slog.Info("Key created", "kid", kid)
}

func createECXKey(curve elliptic.Curve) {
	algName := fmt.Sprintf("ES%s", curve.Params().Name[2:])
	if algName == "ES521" {
		algName = "ES512"
	}
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		slog.Error("Couldn't generate ECDSA key", "error", err)
		os.Exit(1)
	}
	publicKey := privateKey.Public()
	privateJWK, err := jwk.New(privateKey)
	if err != nil {
		slog.Error("Couldn't generate JWK", "error", err)
		os.Exit(1)
	}
	publicJWK, err := jwk.New(publicKey)
	if err != nil {
		slog.Error("Couldn't generate JWK", "error", err)
		os.Exit(1)
	}
	kidHash, err := publicJWK.Thumbprint(crypto.SHA256)
	if err != nil {
		slog.Error("Couldn't generate JWK KID", "error", err)
		os.Exit(1)
	}
	kid := fmt.Sprintf("%X", kidHash[0:8])
	if err := privateJWK.Set("alg", algName); err != nil {
		slog.Error("Couldn't set JWK alg", "error", err)
		os.Exit(1)
	}
	if err := publicJWK.Set("alg", algName); err != nil {
		slog.Error("Couldn't set JWK alg", "error", err)
		os.Exit(1)
	}
	if err := privateJWK.Set("use", "sig"); err != nil {
		slog.Error("Couldn't set JWK use", "error", err)
		os.Exit(1)
	}
	if err := publicJWK.Set("use", "sig"); err != nil {
		slog.Error("Couldn't set JWK use", "error", err)
		os.Exit(1)
	}
	if err := privateJWK.Set("kid", kid); err != nil {
		slog.Error("Couldn't set JWK kid", "error", err)
		os.Exit(1)
	}
	if err := publicJWK.Set("kid", kid); err != nil {
		slog.Error("Couldn't set JWK kid", "error", err)
		os.Exit(1)
	}

	privateString, _ := json.Marshal(privateJWK)
	publicString, _ := json.Marshal(publicJWK)

	tx := db.MustBegin()
	if _, err := tx.Exec(`call create_server_key(?, ?, ?, ?, ?)`, kid, "EC", algName, privateString, publicString); err != nil {
		slog.Error("Couldn't create server key", "error", err)
		os.Exit(1)
	}
	if err := tx.Commit(); err != nil {
		slog.Error("Couldn't commit transaction", "error", err)
		os.Exit(1)
	}
	slog.Info("Key created", "kid", kid)
}

func Main(cmd *cobra.Command, args []string) {
	var err error
	db, err = gindb.Connect("mysql", viper.GetString("database.dsn"))
	if err != nil {
		slog.Error("Couldn't connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("Mysql", "dsn", viper.GetString("database.dsn"))
	switch *KeyAlg {
	case "RS256":
		createRSXKey(256)
	case "RS384":
		createRSXKey(384)
	case "RS512":
		createRSXKey(512)
	case "ES256":
		createECXKey(elliptic.P256())
	case "ES384":
		createECXKey(elliptic.P384())
	case "ES512":
		createECXKey(elliptic.P521())
	default:
		slog.Error("Unsupported algorithm", "alg", *KeyAlg)
		os.Exit(1)
	}
}
