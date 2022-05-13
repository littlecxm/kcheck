package checksum

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/littlecxm/kcheck/configs"
	"github.com/littlecxm/kcheck/pkg/fileutil"
	"hash"
	"io"
	"os"
	"path/filepath"
)

func CheckByHash(relativePath, hashStr string, hasher hash.Hash) error {
	defer func() {
		hasher.Reset()
	}()

	fPath := filepath.Join(configs.WorkDir, relativePath)
	if !fileutil.FileExists(fPath) {
		return errors.New("NOT FOUND")
	}

	// path/to/whatever exists
	f, err := os.Open(fPath)
	if err != nil {
		return errors.New("OPEN FAIL")
	}
	defer func() {
		_ = f.Close()
	}()

	if _, err := io.Copy(hasher, f); err != nil {
		return errors.New("failed copy buffer")
	}

	byteHash, _ := hex.DecodeString(hashStr)
	if bytes.Compare(hasher.Sum(nil), byteHash) != 0 {
		return errors.New("CHECK FAIL")
	}
	return nil
}
