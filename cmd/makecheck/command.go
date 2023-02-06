package main

import (
	"crypto/sha256"
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
	builtVer := fmt.Sprintf("makecheck %s %s (built: %s)", version, commitID, buildDate)
	fmt.Println(builtVer)
	fmt.Println("--------")

	if c.NArg() == 0 {
		cli.ShowAppHelpAndExit(c, 0)
	}
	if c.NArg() > 0 {
		srcDir = c.Args().Get(0)
	}

	var kcheckList configs.KCheckList
	res := make(chan *reporter.CheckResult, 999)

	go reporter.Handler("failed_make.list", res)

	kcheckList.CreatedAt = time.Now().Unix()
	kcheckList.Version = builtVer

	// file walk
	var fCount, passCount, failCount = 0, 0, 0
	err := filepath.Walk(srcDir, func(filePath string, fi os.FileInfo, err error) error {
		if err != nil {
			failCount++
			return err
		}
		if utils.CheckPathBlacklist(filePath) {
			return nil
		}
		srcPath := filepath.ToSlash(strings.TrimPrefix(strings.TrimPrefix(filePath, srcDir), string(filepath.Separator)))
		absPath := filepath.Join(srcDir, srcPath)
		if !fi.Mode().IsRegular() {
			return nil
		}
		fCount++

		f, err := os.Open(absPath)
		if err != nil {
			return err
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			res <- &reporter.CheckResult{
				Error: err,
				Path:  srcPath,
			}
			failCount++
			utils.PrintStatus(false, srcPath)
			return err
		} else {
			passCount++
			utils.PrintStatus(true, srcPath)
		}
		kcheckList.Files = append(kcheckList.Files,
			configs.KCheckFiles{
				Path:   srcPath,
				SHA256: hex.EncodeToString(h.Sum(nil)),
				Size:   fi.Size(),
			},
		)
		f.Close()
		return nil
	})
	if err != nil {
		log.Fatalln("filepath err:", err)
	}

	// save list
	var outBytes []byte
	if isIndent {
		outBytes, err = json.MarshalIndent(kcheckList, "", " ")
	} else {
		outBytes, err = json.Marshal(kcheckList)
	}
	err = os.WriteFile(filepath.Join(configs.WorkDir, outFilename), outBytes, os.ModePerm)
	if err != nil {
		log.Fatalln("save list err:", err)
	}

	if failCount == 0 {
		os.Remove(filepath.Join(configs.WorkDir, "failed_make.list"))
	}

	fmt.Println("--------")
	fmt.Printf(
		"Finished.\n"+
			"ALL: %d / PASS: %d / FAIL: %d\n",
		fCount,
		passCount,
		failCount,
	)
	fmt.Println("Check list saved:", outFilename)
	fmt.Println("Exit after 3 seconds...")
	time.Sleep(3 * time.Second)
	return nil
}
