package challengedb

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gitlab.com/daedaluz/gindb"
	"net/http"
	"net/url"
	"time"
	api2 "uyulala/internal/api"
	"uyulala/internal/db"
)

const (
	StatusPending   = "pending"
	StatusViewed    = "viewed"
	StatusSigned    = "signed"
	StatusCollected = "collected"
	StatusRejected  = "rejected"
)

func CreateChallenge(ctx *gin.Context, typ, appID string, expire time.Time,
	publicData, privateData any, signatureText string, signatureData []byte, redirectURL string) (string, error) {
	var pubData, privData []byte
	var err error
	if pubData, err = db.GobEncodeData(publicData); err != nil {
		return "", err
	}
	if privData, err = db.GobEncodeData(privateData); err != nil {
		return "", err
	}
	tx := gindb.GetTX(ctx)
	res, err := tx.Queryx(`call create_challenge(?, ?, ?, ?, ?, ?, ?, ?, ?)`, db.GenerateID(8), typ, appID, expire,
		pubData, privData,
		signatureText, signatureData,
		redirectURL)
	if err != nil {
		return "", err
	}
	defer res.Close()
	var challengeID string
	if !res.Next() {
		return "", sql.ErrNoRows
	}
	if err := res.Scan(&challengeID); err != nil {
		return "", err
	}
	return challengeID, nil
}

type Data struct {
	Created  time.Time `db:"created"`
	ID       string    `db:"id"`
	Type     string    `db:"type"`
	AppID    string    `db:"app_id"`
	PubData  []byte    `db:"public_data"`
	PrivData []byte    `db:"private_data"`

	SignatureText string `db:"signature_text"`
	SignatureData []byte `db:"signature_data"`

	Signature  []byte       `db:"signature"`
	Credential []byte       `db:"credential"`
	Signed     sql.NullTime `db:"signed"`
	Expire     time.Time    `db:"expire"`

	Status        string `db:"status"`
	RedirectURL   string `db:"redirect_url"`
	OAuth2Context string `db:"oauth2_context"`
}

func (c *Data) GetOAuth2Context() url.Values {
	if c.OAuth2Context == "" {
		return url.Values{}
	}
	if values, err := url.ParseQuery(c.OAuth2Context); err == nil {
		return values
	}
	return url.Values{}
}

func (c *Data) Expand(pubOut, privOut any) (err error) {
	if pubOut != nil {
		if err := db.GobDecodeData(c.PubData, pubOut); err != nil {
			return err
		}
	}
	if privOut != nil {
		if err := db.GobDecodeData(c.PrivData, privOut); err != nil {
			return err
		}
	}
	return
}

func (c *Data) Expired() bool {
	return c.Expire.Before(time.Now())
}

func (c *Data) Validate(ctx *gin.Context) bool {
	if c.Signed.Valid {
		api2.AbortError(ctx, http.StatusBadRequest, "signed", "Challenge has already been signed", nil)
		return false
	}

	if c.Status == StatusRejected {
		api2.AbortError(ctx, http.StatusBadRequest, "rejected", "Challenge has already been rejected", nil)
		return false
	}

	if c.Expired() {
		api2.AbortError(ctx, http.StatusBadRequest, "expired", "Challenge has expired", nil)
		return false
	}
	return true
}

func (c *Data) ValidateCollect(ctx *gin.Context) bool {
	if c.Expired() {
		api2.StatusResponse(ctx, http.StatusBadRequest, "expired", "Challenge has expired")
		return false
	}
	switch c.Status {
	case StatusPending:
		api2.StatusResponse(ctx, http.StatusOK, "pending", "Challenge has not been signed yet")
	case StatusViewed:
		api2.StatusResponse(ctx, http.StatusOK, "viewed", "Challenge has not been signed yet")
	case StatusRejected:
		api2.StatusResponse(ctx, http.StatusOK, "rejected", "Challenge has been rejected")
	case StatusCollected:
		api2.StatusResponse(ctx, http.StatusBadRequest, "collected", "Challenge has already been collected")
	case StatusSigned:
		return true
	default:
		api2.StatusResponse(ctx, http.StatusBadRequest, "invalid_status", "Invalid challenge status")
	}
	return false
}

func GetChallenge(ctx *gin.Context, challengeID string) (ch *Data, err error) {
	ch = &Data{}
	tx := gindb.GetTX(ctx)
	res, err := tx.Queryx(`call get_challenge(?)`, challengeID)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if !res.Next() {
		return nil, sql.ErrNoRows
	}
	if err := res.StructScan(ch); err != nil {
		return nil, err
	}
	return
}

func SignChallenge(ctx *gin.Context, challengeID string, signature *protocol.ParsedCredentialAssertionData, credential *webauthn.Credential) error {
	tx := gindb.GetTX(ctx)

	sig, err := db.GobEncodeData(signature)
	if err != nil {
		return err
	}
	cred, err := db.GobEncodeData(credential)
	if err != nil {
		return err
	}
	if err := SetChallengeStatus(ctx, challengeID, StatusSigned); err != nil {
		return err
	}
	_, err = tx.Exec(`call sign_challenge(?, ?, ?)`, challengeID, sig, cred)
	return err
}

func DeleteChallenge(ctx *gin.Context, challengeID string) error {
	tx := gindb.GetTX(ctx)
	_, err := tx.Exec(`call delete_challenge(?)`, challengeID)
	return err
}

func SetChallengeStatus(ctx *gin.Context, challengeID, status string) error {
	tx := gindb.GetTX(ctx)
	_, err := tx.Exec(`call set_challenge_status(?, ?)`, challengeID, status)
	return err
}

func SetOAuth2Context(ctx *gin.Context, challengeID, context string) error {
	tx := gindb.GetTX(ctx)
	_, err := tx.Exec(`call set_oauth2_context(?, ?)`, challengeID, context)
	return err
}

func GetChallengeByCode(ctx *gin.Context, challengeCode string) (ch *Data, err error) {
	ch = &Data{}
	tx := gindb.GetTX(ctx)
	res, err := tx.Queryx(`call get_challenge_by_code(?)`, challengeCode)
	if err != nil {
		return nil, err
	}
	if !res.Next() {
		return nil, sql.ErrNoRows
	}
	if err := res.StructScan(ch); err != nil {
		return nil, err
	}
	return
}

func CreateCode(ctx *gin.Context, challengeID string) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	tx := gindb.GetTX(ctx)
	_, err = tx.Exec(`call create_code(?, ?)`, id.String(), challengeID)
	return id.String(), err
}

func DeleteCode(ctx *gin.Context, code string) error {
	tx := gindb.GetTX(ctx)
	_, err := tx.Exec(`call delete_code(?)`, code)
	return err
}
