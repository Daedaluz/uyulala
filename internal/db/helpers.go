package db

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"

	"github.com/google/uuid"
)

func GobEncodeData(data any) ([]byte, error) {
	buff := &bytes.Buffer{}
	enc := gob.NewEncoder(buff)
	if err := enc.Encode(data); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func GobDecodeData(data []byte, dest any) error {
	buff := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buff)
	if err := dec.Decode(dest); err != nil {
		return err
	}
	return nil
}

func GenerateUUID() string {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return id.String()
}

func GenerateID(length int) string {
	data := make([]byte, length)
	_, _ = rand.Read(data)
	uid := hex.EncodeToString(data)
	return uid
}
