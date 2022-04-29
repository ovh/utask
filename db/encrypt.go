package db

import (
	"sync"

	"github.com/loopfz/gadgeto/zesty"
)

type KeyRotationCallback func(dbp zesty.DBProvider) error

var (
	keyRotationsCb   []KeyRotationCallback
	keyRotationsCbMu sync.Mutex
)

// RegisterKeyRotations registers a callback which will be called during the encrypt
// keys rotations
func RegisterKeyRotations(cb KeyRotationCallback) {
	keyRotationsCbMu.Lock()
	defer keyRotationsCbMu.Unlock()

	keyRotationsCb = append(keyRotationsCb, cb)
}

// CallKeyRotations calls registered callbacks to rotate the encryption keys
func CallKeyRotations(dbp zesty.DBProvider) error {
	keyRotationsCbMu.Lock()
	defer keyRotationsCbMu.Unlock()

	for _, cb := range keyRotationsCb {
		if err := cb(dbp); err != nil {
			return err
		}
	}

	return nil
}
