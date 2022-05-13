package filetype

import (
	"bufio"
	"encoding/binary"
	"errors"
	"github.com/littlecxm/kcheck/configs"
	"github.com/littlecxm/kcheck/pkg/fileutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// getDefaultPaths are return file list paths
func getDefaultPaths() []string {
	return []string{
		"allfiles.lst",
		filepath.Join("data", "__metadata.metatxt"),
		filepath.Join("prop", "filepath.xml"),
	}
}

// GuessListPath get the relative path of the list file
func GuessListPath() (string, error) {
	for _, p := range getDefaultPaths() {
		if fileutil.FileExists(p) {
			return p, nil
		}
	}
	return "", errors.New("failed to guess input list, please specify the path")
}

func GuessType(list string) (string, error) {
	file, err := os.Open(filepath.Join(configs.WorkDir, list))
	buff := bufio.NewReader(file)

	// check like kbin
	magicNumber := make([]byte, 2)
	if _, err := buff.Read(magicNumber); err != nil {
		return "", err
	}
	if binary.BigEndian.Uint16(magicNumber) == configs.KBinMagicNumber {
		return configs.KBinType, nil
	}

	// detect type
	rb, err := ioutil.ReadFile(list)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(string(rb), "<?xml ") {
		return configs.XMLType, nil
	}
	if strings.Contains(string(rb), "createdAt") {
		return configs.MetadataType, nil
	}

	// default
	return configs.KCheckType, nil
}
