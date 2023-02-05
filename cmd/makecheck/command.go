package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/littlecxm/kcheck/configs"
	"github.com/littlecxm/kcheck/pkg/reporter"
	"github.com/littlecxm/kcheck/pkg/utils"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func commandHandler(c *cli.Context) error {
	fmt.Sprintf("makecheck %s %s (built: %s)", version, commitID, buildDate)

	if c.NArg() == 0 {
		cli.ShowAppHelpAndExit(c, 0)
	}
	if c.NArg() > 0 {
		srcDir = c.Args().Get(0)
	}

	h := sha1.New()
	var kcheckList configs.KCheckList
	res := make(chan *reporter.CheckResult, 999)

	go reporter.Handler("failed_make.list", res)

	err := filepath.Walk(srcDir, func(filePath string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(fi.Name()), ".ds_store") {
			return nil
		}
		srcPath := filepath.ToSlash(strings.TrimPrefix(strings.TrimPrefix(filePath, srcDir), string(filepath.Separator)))
		if !fi.Mode().IsRegular() {
			return nil
		}
		f, err := os.Open(filePath)
		_, err = io.Copy(h, f)
		if err != nil {
			res <- &reporter.CheckResult{
				Error: err,
				Path:  srcPath,
			}
			utils.PrintStatus(false, srcPath)
			return err
		} else {
			utils.PrintStatus(true, srcPath)
		}
		kcheckList.Files = append(kcheckList.Files, configs.KCheckFiles{Path: srcPath, SHA1: hex.EncodeToString(h.Sum(nil)), Size: fi.Size()})
		f.Close()
		return nil
	})
	if err != nil {
		log.Fatalln("filepath err:", err)
	}
	var outBytes []byte
	if isIndent {
		outBytes, err = json.Marshal(kcheckList)
	} else {
		outBytes, err = json.MarshalIndent(kcheckList, "", " ")
	}
	err = os.WriteFile(filepath.Join(configs.WorkDir, outFilename), outBytes, os.ModePerm)
	if err != nil {
		log.Fatalln("save list err:", err)
	}
	fmt.Println("--------")
	fmt.Println("Check list saved:", outFilename)
	fmt.Println("Finished.")
	fmt.Println("Exit after 2 seconds...")
	time.Sleep(3 * time.Second)
	return nil
}
