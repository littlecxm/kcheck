package checksum

import (
	"bytes"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"
	"path/filepath"

	"github.com/littlecxm/kcheck/configs"
	"github.com/littlecxm/kcheck/pkg/utils"
)

func checkByHash(relativePath, hashStr string, newHash func() hash.Hash) error {
	h := newHash()
	if h == nil {
		return errors.New("hash unavailable")
	}

	fPath := filepath.Join(configs.WorkDir, relativePath)
	if !utils.FileExists(fPath) {
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

	if _, err := io.Copy(h, f); err != nil {
		return errors.New("failed copy buffer")
	}

	byteHash, err := hex.DecodeString(hashStr)
	if err != nil {
		return errors.New("cannot decode hash string")
	}
	if bytes.Compare(h.Sum(nil), byteHash) != 0 {
		return errors.New("CHECK FAIL")
	}
	return nil
}
