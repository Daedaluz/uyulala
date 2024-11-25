package mds

import (
	"sync"

	"github.com/go-webauthn/webauthn/metadata"
	"github.com/google/uuid"
)

var (
	lock = sync.Mutex{}
	meta map[uuid.UUID]*metadata.Entry
)

func getMeta() (map[uuid.UUID]*metadata.Entry, error) {
	lock.Lock()
	defer lock.Unlock()
	m := meta
	if m == nil {
		tmp, err := metadata.Fetch()
		if err != nil {
			return nil, err
		}
		m = tmp.ToMap()
		meta = m
	}
	return m, nil
}

func Init() {
	_, _ = getMeta()
}

func Get(aaguid uuid.UUID) (*metadata.Entry, error) {
	m, err := getMeta()
	if err != nil {
		return nil, err
	}
	return m[aaguid], nil
}
