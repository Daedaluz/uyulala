package sessiondb

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/spf13/viper"
	"gitlab.com/daedaluz/gindb"
	"time"
	"uyulala/internal/db"
)

type Session struct {
	ID              string       `db:"id" json:"id"`
	UserID          string       `db:"user_id" json:"userId"`
	AppID           string       `db:"app_id" json:"appId"`
	Counter         uint32       `db:"counter" json:"counter"`
	RequestedScopes string       `db:"requested_scopes" json:"requestedScopes"`
	Created         time.Time    `db:"created_at" json:"created"`
	ExpireAt        sql.NullTime `db:"expire_at" json:"expires"`
}

func Get(c *gin.Context, sessionID string) (*Session, error) {
	s := &Session{}
	tx := gindb.GetTX(c)
	err := tx.Get(s, `call get_session(?)`, sessionID)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func Create(c *gin.Context, userID, appID, scopes string) (*Session, error) {
	dur := viper.GetDuration("refreshToken.length")
	exp := time.Time{}
	if dur != 0 {
		exp = time.Now().Add(dur)
	}
	sess := &Session{
		ID:              db.GenerateID(8),
		UserID:          userID,
		AppID:           appID,
		RequestedScopes: scopes,
		Counter:         0,
		ExpireAt:        sql.NullTime{},
	}
	if !exp.IsZero() {
		sess.ExpireAt = sql.NullTime{
			Time:  exp,
			Valid: true,
		}
	}

	tx := gindb.GetTX(c)
	_, err := tx.Exec(`call create_session(?, ?, ?, ?, ?)`,
		sess.ID, userID, appID, scopes, sess.ExpireAt)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func Delete(c *gin.Context, sid string) error {
	tx := gindb.GetTX(c)
	_, err := tx.Exec(`call delete_session(?)`, sid)
	return err
}

func Rotate(c *gin.Context, session *Session) error {
	tx := gindb.GetTX(c)
	session.Counter++
	if viper.GetBool("refreshToken.extendOnUse") && viper.GetDuration("refreshToken.length") > 0 {
		session.ExpireAt = sql.NullTime{
			Time:  time.Now().Add(viper.GetDuration("refreshToken.length")),
			Valid: true,
		}
	}
	_, err := tx.Exec(`call rotate_session(?, ?)`, session.ID, session.ExpireAt)
	return err
}

// CreateRefreshToken generates a signed jwt token with the session id as the jwt id and a counter-claim to prevent reuse of the token.
func (s *Session) CreateRefreshToken(key jwk.Key) (string, error) {
	hdrs := jws.NewHeaders()
	_ = hdrs.Set("typ", "refresh+jwt")
	token := jwt.New()
	_ = token.Set(jwt.JwtIDKey, s.ID)
	_ = token.Set("counter", s.Counter)
	_ = token.Set(jwt.IssuedAtKey, time.Now())
	_ = token.Set(jwt.IssuerKey, viper.GetString("issuer"))
	_ = token.Set(jwt.AudienceKey, viper.GetString("issuer"))
	_ = token.Set(jwt.SubjectKey, s.AppID)

	tokenBytes, err := jwt.Sign(token, jwa.SignatureAlgorithm(key.Algorithm()), key, jwt.WithHeaders(hdrs))
	if err != nil {
		return "", err
	}
	return string(tokenBytes), nil
}

func ListForUser(c *gin.Context, userID string) ([]*Session, error) {
	tx := gindb.GetTX(c)
	var sessions []*Session
	err := tx.Select(&sessions, `call list_sessions_for_user(?)`, userID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}
