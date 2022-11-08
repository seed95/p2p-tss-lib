package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/tss"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// To change these parameters, you must first run the `TestDoKeyGenAndSaveKeys` in keygen_test alone to save keys
// with new Participants or Threshold.
const (
	Participants = 2
	Threshold    = 1
)

const (
	keyDirFormat  = "%s/keys"
	keyFileFormat = "data_%d.json"
)

func SaveTestKey(t *testing.T, index int, key keygen.LocalPartySaveData) {
	keyDir := getKeyDirectory()
	keyFileName := fmt.Sprintf("%s/"+keyFileFormat, keyDir, index)

	// If the key is already saved, do nothing.
	if _, err := os.Stat(keyFileName); err == nil {
		return
	}

	keyFile, err := os.Create(keyFileName)
	require.NoErrorf(t, err, "unable to create file %s", keyFileName)

	bz, err := json.Marshal(&key)
	require.NoErrorf(t, err, "unable to marshal key for save in file %s", keyFileName)

	_, err = keyFile.Write(bz)
	require.NoErrorf(t, err, "unable to write to file %s", keyFileName)
}

func LoadTestKeys(numParty int) ([]keygen.LocalPartySaveData, error) {
	if numParty > Participants {
		return nil, errors.New("invalid num party")
	}

	keyDir := getKeyDirectory()
	keys := make([]keygen.LocalPartySaveData, 0, numParty)
	for i := 0; i < numParty; i++ {
		filePath := fmt.Sprintf("%s/"+keyFileFormat, keyDir, i)
		bz, err := ioutil.ReadFile(filePath)
		if err != nil {
			return nil, errors.Wrapf(err, "could not open the test key for party %d in the expected location: %s.", i, filePath)
		}
		var key keygen.LocalPartySaveData
		if err = json.Unmarshal(bz, &key); err != nil {
			return nil, errors.Wrapf(err, "could not unmarshal key data for party %d located at: %s", i, filePath)
		}
		for _, kbxj := range key.BigXj {
			kbxj.SetCurve(tss.S256())
		}
		key.ECDSAPub.SetCurve(tss.S256())
		keys = append(keys, key)
	}

	return keys, nil
}

func RemoveKeysAndMakeDir(t *testing.T) {
	keyDir := getKeyDirectory()

	// Remove keys
	err := os.RemoveAll(keyDir)
	require.NoError(t, err, "unable to remove key directory %s", keyDir)

	// Create new key directory
	err = os.Mkdir(keyDir, 0755)
	require.NoError(t, err, "unable to create key directory %s", keyDir)
}

func getKeyDirectory() string {
	_, callerFileName, _, _ := runtime.Caller(0)
	srcDirName := filepath.Dir(callerFileName)
	keyDir := fmt.Sprintf(keyDirFormat, srcDirName)
	return keyDir
}
