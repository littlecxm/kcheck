package main

import (
	"bufio"
	"encoding/binary"
	"github.com/littlecxm/kcheck/configs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func guessPath() string {
	metaDefaultPath := filepath.Join(workDir, "data", "__metadata.metatxt")
	kbinDefaultPath := filepath.Join(workDir, "prop", "filepath.xml")
	if _, err := os.Stat(metaDefaultPath); err == nil {
		log.Println("use default metadata list:", metaDefaultPath)
		return metaDefaultPath
	}
	if _, err := os.Stat(kbinDefaultPath); err == nil {
		log.Println("use default filepath list:", kbinDefaultPath)
		return kbinDefaultPath
	}
	return ""
}

func guessType() (string, error) {
	file, err := os.Open(filepath.Join(configs.WorkDir, listPath))
	bufr := bufio.NewReader(file)
	// kbin
	magicNumber := make([]byte, 2)
	_, err = bufr.Read(magicNumber)
	if err != nil {
		return "", err
	}
	if binary.BigEndian.Uint16(magicNumber) == configs.KBinMagicNumber {
		return configs.KBinType, nil
	}
	// ioutil
	bf, err := ioutil.ReadFile(listPath)
	if err != nil {
		return "", err
	}
	if strings.Contains(string(bf), "<?xml version") {
		return configs.XMLType, nil
	}
	if strings.Contains(string(bf), "createdAt") {
		return configs.MetadataType, nil
	}
	log.Println("unknown file type, use default:", configs.KCheckType)
	return configs.KCheckType, nil
}
