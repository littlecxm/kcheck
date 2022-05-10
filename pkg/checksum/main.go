package checksum

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"github.com/littlecxm/kcheck/configs"
	"io"
	"os"
	"path/filepath"
)

func CompareFileMD5(relativePath, filemd5 string) error {
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
	h := md5.New()
	io.Copy(h, f)
	//return strings.ToUpper(hex.EncodeToString(h.Sum(nil))), nil
	bytemd5, _ := hex.DecodeString(filemd5)
	if bytes.Compare(h.Sum(nil), bytemd5) != 0 {
		return errors.New("CHECK FAIL")
	}
	return nil
}

func CompareFileSHA1(relativePath, filesha1 string) error {
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
	h := sha1.New()
	io.Copy(h, f)
	//return strings.ToUpper(hex.EncodeToString(h.Sum(nil))), nil
	bytemd5, _ := hex.DecodeString(filesha1)
	if bytes.Compare(h.Sum(nil), bytemd5) != 0 {
		return errors.New("CHECK FAIL")
	}
	return nil
}
