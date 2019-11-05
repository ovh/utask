package models

import (
	"github.com/ovh/configstore"
	"github.com/ovh/symmecrypt"
	"github.com/ovh/symmecrypt/keyloader"
)

// EncryptionKey holds the global key to encrypt/decrypt
// task data in DB
var EncryptionKey symmecrypt.Key

// Init takes an instance of configstore and loads EncryptionKey from it
func Init(store *configstore.Store) error {
	k, err := keyloader.LoadKeyFromStore("storage", store)
	if err != nil {
		return err
	}
	EncryptionKey = k
	return nil
}
