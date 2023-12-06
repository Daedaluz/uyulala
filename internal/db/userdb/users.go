package userdb

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gitlab.com/daedaluz/gindb"
	"time"
	"uyulala/internal/db"
)

func CreateUser(ctx *gin.Context) (string, error) {
	tx := gindb.GetTX(ctx)
	id := db.GenerateID(18)

	res, err := tx.Queryx(`call create_user(?)`, id)
	if err != nil {
		return "", err
	}
	defer res.Close()
	var userID string
	if !res.Next() {
		return "", sql.ErrNoRows
	}
	if err := res.Scan(&userID); err != nil {
		return "", err
	}
	return userID, nil
}

type User struct {
	ID       string    `db:"id"`
	LastAuth time.Time `db:"last_auth"`
	Created  time.Time `db:"created"`
}

func GetUser(ctx *gin.Context, userID string) (*User, error) {
	tx := gindb.GetTX(ctx)
	res, err := tx.Queryx(`call get_user(?)`, userID)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var user User
	if !res.Next() {
		return nil, sql.ErrNoRows
	}
	if err := res.StructScan(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func ListUsers(ctx *gin.Context) ([]string, error) {
	tx := gindb.GetTX(ctx)
	res, err := tx.Queryx(`call list_users()`)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var users []string
	for res.Next() {
		var userID string
		if err := res.Scan(&userID); err != nil {
			return nil, err
		}
		users = append(users, userID)
	}
	return users, nil
}

type UserKey struct {
	Hash     string              `json:"hash"`
	Key      webauthn.Credential `json:"key"`
	AAGuid   string              `json:"aaguid"`
	Created  time.Time           `json:"created"`
	LastUsed time.Time           `json:"lastUsed"`
}
type UserWithKeys struct {
	ID          string    `json:"id"`
	Created     time.Time `json:"created"`
	Credentials []UserKey `json:"keys"`
}

type dbUserWithKeys struct {
	UserID        string       `db:"userId"`
	Created       time.Time    `db:"userCreated"`
	KeyHash       *string      `db:"keyHash"`
	KeyID         []byte       `db:"keyId"`
	KeyAAGUID     *string      `db:"keyAAGUID"`
	KeyCreated    sql.NullTime `db:"keyCreated"`
	KeyLastUsed   sql.NullTime `db:"keyUsed"`
	KeyCredential []byte       `db:"keyCredential"`
}

func GetUserWithKeys(ctx *gin.Context, userID string) (*UserWithKeys, error) {
	tx := gindb.GetTX(ctx)
	res, err := tx.Queryx(`call get_user_with_keys(?)`, userID)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var user *UserWithKeys
	tmp := &UserWithKeys{}
	for res.Next() {
		ent := &dbUserWithKeys{}
		if err := res.StructScan(ent); err != nil {
			return nil, err
		}
		if tmp.ID != ent.UserID {
			if tmp.ID != "" {
				user = tmp
			}
			if tmp.Credentials == nil {
				tmp.Credentials = []UserKey{}
			}
			tmp = &UserWithKeys{}
		}
		tmp.ID = ent.UserID
		tmp.Created = ent.Created
		if ent.KeyID != nil {
			cred := &webauthn.Credential{}
			if err := db.GobDecodeData(ent.KeyCredential, cred); err != nil {
				return nil, err
			}
			hash := sha256.Sum256(ent.KeyID)
			hashStr := hex.EncodeToString(hash[:])
			tmp.Credentials = append(tmp.Credentials, UserKey{
				Hash:     hashStr,
				Key:      *cred,
				AAGuid:   *ent.KeyAAGUID,
				Created:  ent.KeyCreated.Time,
				LastUsed: ent.KeyLastUsed.Time,
			})
		}
	}
	if tmp.ID != "" {
		user = tmp
		if tmp.Credentials == nil {
			tmp.Credentials = []UserKey{}
		}
	}
	return user, nil
}

func ListUsersWithKeys(ctx *gin.Context) ([]*UserWithKeys, error) {
	tx := gindb.GetTX(ctx)
	res, err := tx.Queryx(`call list_users_with_keys()`)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var users []*UserWithKeys
	tmp := &UserWithKeys{}
	for res.Next() {
		ent := &dbUserWithKeys{}
		if err := res.StructScan(ent); err != nil {
			return nil, err
		}
		if tmp.ID != ent.UserID {
			if tmp.ID != "" {
				users = append(users, tmp)
			}
			if tmp.Credentials == nil {
				tmp.Credentials = []UserKey{}
			}
			tmp = &UserWithKeys{}
		}
		tmp.ID = ent.UserID
		tmp.Created = ent.Created
		if ent.KeyID != nil {
			cred := &webauthn.Credential{}
			if err := db.GobDecodeData(ent.KeyCredential, cred); err != nil {
				return nil, err
			}
			hash := sha256.Sum256(ent.KeyID)
			hashStr := hex.EncodeToString(hash[:])
			tmp.Credentials = append(tmp.Credentials, UserKey{
				Hash:     hashStr,
				Key:      *cred,
				AAGuid:   *ent.KeyAAGUID,
				Created:  ent.KeyCreated.Time,
				LastUsed: ent.KeyLastUsed.Time,
			})
		}
	}
	if tmp.ID != "" {
		users = append(users, tmp)
		if tmp.Credentials == nil {
			tmp.Credentials = []UserKey{}
		}
	}
	return users, nil
}

type Key struct {
	Hash       string       `db:"hash"`
	ID         []byte       `db:"id"`
	UserID     string       `db:"user_id"`
	AAGUID     string       `db:"aaguid"`
	Credential []byte       `db:"credential"`
	Created    time.Time    `db:"created"`
	LastUsed   sql.NullTime `db:"last_used"`
}

func GetUserKeyDescriptors(ctx *gin.Context, userID string) ([]protocol.CredentialDescriptor, error) {
	tx := gindb.GetTX(ctx)

	var keys []protocol.CredentialDescriptor

	res, err := tx.Queryx(`call get_user_keys(?)`, userID)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var key Key
		if err := res.StructScan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, protocol.CredentialDescriptor{
			Type:         "public-key",
			CredentialID: key.ID,
		})
	}
	return keys, nil
}

func GetUserKeys(ctx *gin.Context, userID string) []webauthn.Credential {
	tx := gindb.GetTX(ctx)
	var keys []webauthn.Credential
	res, err := tx.Queryx(`call get_user_keys(?)`, userID)
	if err != nil {
		return nil
	}
	defer res.Close()
	for res.Next() {
		var key Key
		if err := res.StructScan(&key); err != nil {
			return nil
		}
		cred := &webauthn.Credential{}
		if err := db.GobDecodeData(key.Credential, cred); err != nil {
			return nil
		}
		keys = append(keys, *cred)
	}
	return keys
}

func GetKey(ctx *gin.Context, keyID []byte) (*Key, error) {
	tx := gindb.GetTX(ctx)
	hash := sha256.Sum256(keyID)
	hashStr := hex.EncodeToString(hash[:])
	res, err := tx.Queryx(`call get_key(?)`, hashStr)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var key Key
	if !res.Next() {
		return nil, sql.ErrNoRows
	}
	if err := res.StructScan(&key); err != nil {
		return nil, err
	}
	return &key, nil
}

func CreateUserKey(ctx *gin.Context, userID string, aaguid uuid.UUID, credential *webauthn.Credential) error {
	tx := gindb.GetTX(ctx)
	cred, err := db.GobEncodeData(credential)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(credential.ID)
	hashStr := hex.EncodeToString(hash[:])
	_, err = tx.Exec(`call add_user_key(?,?,?,?,?)`, hashStr, credential.ID, aaguid.String(), userID, cred)
	return err
}

func PingUserKey(ctx *gin.Context, cred *webauthn.Credential) error {
	tx := gindb.GetTX(ctx)

	hash := sha256.Sum256(cred.ID)
	hashStr := hex.EncodeToString(hash[:])
	credData, err := db.GobEncodeData(cred)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`call ping_user_key(?, ?)`, hashStr, credData)
	return err
}

func DeleteUser(ctx *gin.Context, userID string) error {
	tx := gindb.GetTX(ctx)
	_, err := tx.Exec(`call delete_user(?)`, userID)
	return err
}

func DeleteUserKey(ctx *gin.Context, userID, keyHash string) error {
	tx := gindb.GetTX(ctx)
	_, err := tx.Exec(`call delete_user_key(?,?)`, userID, keyHash)
	return err
}

func UpdateAuthTime(ctx *gin.Context, userID, appID string) error {
	tx := gindb.GetTX(ctx)
	_, err := tx.Exec(`call update_user_auth_time(?, ?)`, userID, appID)
	return err
}

func GetAuthTime(ctx *gin.Context, userID, appID string) (time.Time, error) {
	tx := gindb.GetTX(ctx)
	var authTime time.Time
	err := tx.Get(&authTime, `call get_user_auth_time(?,?)`, userID, appID)
	if err != nil {
		return time.Time{}, err
	}
	return authTime, nil
}
