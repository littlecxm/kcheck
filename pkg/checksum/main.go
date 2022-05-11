package checksum

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/littlecxm/kcheck/configs"
	"hash"
	"io"
	"os"
	"path/filepath"
)

func CheckByHash(relativePath, hashStr string, hasher hash.Hash) error {
	fPath := filepath.Join(configs.WorkDir, relativePath)
	if _, err := os.Stat(fPath); err != nil {
		return errors.New("NOT FOUND")
	}
	// path/to/whatever exists
	f, err := os.Open(fPath)
	if err != nil {
		return errors.New("OPEN FAIL")
	}
	defer f.Close()
	h := hasher
	io.Copy(h, f)
	//return strings.ToUpper(hex.EncodeToString(h.Sum(nil))), nil
	byteHash, _ := hex.DecodeString(hashStr)
	if bytes.Compare(h.Sum(nil), byteHash) != 0 {
		return errors.New("CHECK FAIL")
	}
	return nil
}
