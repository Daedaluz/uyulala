package challengedb

import (
	"database/sql"
	"net/http"
	"net/url"
	"time"
	"uyulala/internal/api"
	"uyulala/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gitlab.com/daedaluz/gindb"
)

const (
	StatusPending   = "pending"
	StatusViewed    = "viewed"
	StatusSigned    = "signed"
	StatusCollected = "collected"
	StatusRejected  = "rejected"
)

func CreateChallenge(ctx *gin.Context, typ, appID string, expire time.Time,
	publicData, privateData any, signatureText string, signatureData []byte, redirectURL string) (string, string, error) {
	var pubData, privData []byte
	var err error
	var secret uuid.UUID
	secret, err = uuid.NewRandom()

	if err != nil {
		return "", "", err
	}
	if pubData, err = db.GobEncodeData(publicData); err != nil {
		return "", "", err
	}
	if privData, err = db.GobEncodeData(privateData); err != nil {
		return "", "", err
	}

	tx := gindb.GetTX(ctx)
	res, err := tx.Queryx(`call create_challenge(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, db.GenerateID(8), typ, appID, expire,
		pubData, privData,
		signatureText, signatureData,
		redirectURL, secret)
	if err != nil {
		return "", "", err
	}
	defer res.Close()
	var challengeID string
	if !res.Next() {
		return "", "", sql.ErrNoRows
	}
	if err := res.Scan(&challengeID); err != nil {
		return "", "", err
	}
	return challengeID, secret.String(), nil
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
	Secret        string `db:"secret"`
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
		api.AbortError(ctx, http.StatusBadRequest, "signed", "Challenge has already been signed", nil)
		return false
	}

	if c.Status == StatusRejected {
		api.AbortError(ctx, http.StatusBadRequest, "access_denied", "Challenge has already been rejected", nil)
		return false
	}

	if c.Expired() {
		api.AbortError(ctx, http.StatusBadRequest, "expired_token", "Challenge has expired", nil)
		return false
	}
	return true
}

func (c *Data) ValidateBIDCollect(ctx *gin.Context) bool {
	if c.Expired() {
		api.StatusResponse(ctx, http.StatusBadRequest, "expired", "Challenge has expired")
		return false
	}
	switch c.Status {
	case StatusPending:
		api.StatusResponse(ctx, http.StatusBadRequest, "pending", "Waiting for user to view the challenge")
	case StatusViewed:
		api.StatusResponse(ctx, http.StatusBadRequest, "viewed", "Waiting for user to sign the challenge")
	case StatusRejected:
		api.StatusResponse(ctx, http.StatusBadRequest, "rejected", "Challenge has been rejected")
	case StatusCollected:
		api.StatusResponse(ctx, http.StatusBadRequest, "collected", "Challenge has already been collected")
	case StatusSigned:
		return true
	default:
		api.StatusResponse(ctx, http.StatusInternalServerError, "invalid_status", "Invalid challenge status")
	}
	return false
}
func (c *Data) ValidateOAuthCollect(ctx *gin.Context) bool {
	if c.Expired() {
		api.OAuth2ErrorResponse(ctx, http.StatusBadRequest, "expired_token", "Challenge has expired")
		return false
	}
	switch c.Status {
	case StatusPending:
		api.OAuth2ErrorResponse(ctx, http.StatusBadRequest, "authorization_pending", "Waiting for user to view the challenge")
	case StatusViewed:
		api.OAuth2ErrorResponse(ctx, http.StatusBadRequest, "authorization_viewed", "Waiting for user to sign the challenge")
	case StatusRejected:
		api.OAuth2ErrorResponse(ctx, http.StatusBadRequest, "access_denied", "Challenge has been rejected")
	case StatusCollected:
		api.OAuth2ErrorResponse(ctx, http.StatusBadRequest, "collected", "Challenge has already been collected")
	case StatusSigned:
		return true
	default:
		api.OAuth2ErrorResponse(ctx, http.StatusInternalServerError, "invalid_status", "Invalid challenge status")
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

func SignCreationChallenge(ctx *gin.Context, challengeID string, signature *protocol.ParsedCredentialCreationData, credential *webauthn.Credential) error {
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
	res := tx.QueryRowx(`call get_challenge_by_code(?)`, challengeCode)
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
	x, err := tx.Exec(`call delete_code(?)`, code)
	if err != nil {
		return err
	}
	n, err := x.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}

func CreateCIBARequestID(ctx *gin.Context, challengeID string) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	tx := gindb.GetTX(ctx)
	_, err = tx.Exec(`call create_ciba_request_id(?, ?)`, id.String(), challengeID)
	return id.String(), err
}

func GetChallengeByCIBARequestID(ctx *gin.Context, requestID string) (ch *Data, err error) {
	ch = &Data{}
	tx := gindb.GetTX(ctx)
	res := tx.QueryRowx(`call get_challenge_by_ciba_request_id(?)`, requestID)
	if err := res.StructScan(ch); err != nil {
		return nil, err
	}
	return
}

func DeleteCIBARequest(ctx *gin.Context, requestID string) error {
	tx := gindb.GetTX(ctx)
	x, err := tx.Exec(`call delete_ciba_request(?)`, requestID)
	if err != nil {
		return err
	}
	n, err := x.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return sql.ErrNoRows
	}
	return nil
}
