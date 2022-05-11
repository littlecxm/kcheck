package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/littlecxm/kcheck/configs"
	"github.com/littlecxm/kcheck/pkg/utils"
	"github.com/urfave/cli/v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var (
	version, buildDate, commitID string
	srcDir, outFilename          string
	isIndent                     bool
)

func main() {
	fmt.Printf("makecheck %s\n", version)
	configs.WorkDir, _ = os.Getwd()
	app := &cli.App{
		Name:    "makecheck",
		Usage:   "make check list for kcheck",
		Version: fmt.Sprintf("%s %s (built: %s)", version, commitID, buildDate),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "src",
				Aliases:     []string{"s"},
				Usage:       "source `DIR`",
				Destination: &srcDir,
			},
			&cli.StringFlag{
				Name:        "output",
				Aliases:     []string{"o"},
				Value:       "kcheck.json",
				Usage:       "output filename for check list `FILE`",
				Destination: &outFilename,
			},
			&cli.BoolFlag{
				Name:        "indent",
				Aliases:     []string{"ind"},
				Value:       false,
				Usage:       "enable indent for output list",
				Destination: &isIndent,
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				cli.ShowAppHelpAndExit(c, 0)
			}
			if c.NArg() > 0 {
				srcDir = c.Args().Get(0)
			}

			h := sha1.New()
			var kcheckList configs.KCheckList
			res := make(chan *CheckResult, 999)
			go handler(res)
			err := filepath.Walk(srcDir, func(filePath string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if strings.Contains(strings.ToLower(fi.Name()), ".ds_store") {
					return nil
				}
				srcPath := filepath.ToSlash(strings.TrimPrefix(strings.Replace(filePath, srcDir, "", -1), string(filepath.Separator)))
				if !fi.Mode().IsRegular() {
					return nil
				}
				f, err := os.Open(filePath)
				_, err = io.Copy(h, f)
				if err != nil {
					res <- &CheckResult{
						false,
						err,
						srcPath,
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
			err = ioutil.WriteFile(filepath.Join(configs.WorkDir, outFilename), outBytes, os.ModePerm)
			if err != nil {
				log.Fatalln("save list err:", err)
			}
			fmt.Println("--------")
			fmt.Println("Check list saved:", outFilename)
			fmt.Println("Finished.")
			fmt.Println("Exit after 2 seconds...")
			time.Sleep(3 * time.Second)
			return nil
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
