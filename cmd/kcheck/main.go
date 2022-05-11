package main

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"github.com/beevik/etree"
	"github.com/fatih/color"
	jsoniter "github.com/json-iterator/go"
	"github.com/littlecxm/kbinxml-go"
	"github.com/littlecxm/kcheck/configs"
	"github.com/littlecxm/kcheck/pkg/checksum"
	"github.com/urfave/cli/v2"
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
	workDir                      string
	version, buildDate, commitID string
	listType, listPath           string
)

func main() {
	fmt.Printf("kcheck %s\n", version)
	configs.WorkDir, _ = os.Getwd()
	app := &cli.App{
		Name:    "kcheck",
		Usage:   "check files through list",
		Version: fmt.Sprintf("%s (built: %s)", commitID, buildDate),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "type",
				Aliases:     []string{"t"},
				Usage:       "input list `TYPE`, support: `kbin`,`xml`,`metadata`,`kcheck`",
				Destination: &listType,
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				listPath = guessListPath()
				if listPath == "" {
					fmt.Println("failed to guess input list, please specify the path")
					fmt.Fprintln(color.Output, "use",
						color.GreenString("help"),
						"to get more info",
					)
					os.Exit(-1)
				}
			}
			if c.NArg() > 0 {
				manualPath := c.Args().Get(0)
				listPath = filepath.Base(manualPath)
				configs.WorkDir = filepath.Dir(manualPath)
			}
			log.Println("current work dir:", configs.WorkDir)
			log.Println("list path:", listPath)

			f, err := os.Open(filepath.Join(configs.WorkDir, listPath))
			inByte, err := ioutil.ReadAll(f)
			if err != nil {
				log.Fatalf("load file error: %s", err)
			}
			if listType == "" {
				listType, err = guessType()
				if err != nil {
					log.Fatalf("get list type error: %s", err)
				}
			}

			// report handler
			res := make(chan *CheckResult, 999)
			go handler(res)
			var fCount, passCount, failCount = 0, 0, 0
			switch listType {
			case configs.KBinType:
				kxml, _, err := kbinxml.DeserializeKbin(inByte)
				if err != nil {
					log.Fatalf("cannot unmarshal kbin: %s", err)
				}
				inByte = kxml
				fallthrough
			case configs.XMLType:
				doc := etree.NewDocument()
				if err := doc.ReadFromBytes(inByte); err != nil {
					log.Fatal("load xml list err:", err)
				}
				// check
				listNode := doc.FindElement("list")
				fCount = len(listNode.ChildElements())
				for _, fNode := range listNode.SelectElements("file") {
					dstPath := fNode.SelectElement("dst_path").Text()
					dstMd5 := fNode.SelectElement("dst_md5").Text()
					formatPath := strings.TrimPrefix(filepath.FromSlash(dstPath), string(os.PathSeparator))
					if err := checksum.CheckByHash(formatPath, dstMd5, md5.New()); err != nil {
						failCount++
						res <- &CheckResult{
							false,
							err,
							formatPath,
						}
						printStatus(false, formatPath)
					} else {
						passCount++
						printStatus(true, formatPath)
					}
				}
				close(res)
			case configs.MetadataType:
				var metaStruct configs.MetaData
				err = json.Unmarshal(inByte, &metaStruct)
				if err != nil {
					log.Fatal("load metadata list err:", err)
				}
				metaCreateAt := time.Unix(0, metaStruct.CreatedAt*int64(time.Millisecond))
				fCount = len(metaStruct.Files)
				log.Println("metadata created at:", metaCreateAt)
				for _, files := range metaStruct.Files {
					var (
						fileSHA1 = files.SHA1
						filePath = files.Path
					)
					if fileSHA1 == "" {
						fileSHA1 = files.SSHA1
					}
					if filePath == "" {
						filePath = files.SPath
					}
					formatPath := filepath.Join(
						"data",
						strings.TrimPrefix(filepath.FromSlash(filePath), string(os.PathSeparator)),
					)
					if err := checksum.CheckByHash(formatPath, fileSHA1, sha1.New()); err != nil {
						failCount++
						res <- &CheckResult{
							false,
							err,
							formatPath,
						}
						printStatus(false, formatPath)
					} else {
						passCount++
						printStatus(true, formatPath)
					}
				}
			case configs.KCheckType:
				var kcheckList configs.KCheckList
				err = json.Unmarshal(inByte, &kcheckList)
				if err != nil {
					log.Fatal("load KCheck list err:", err)
				}
				metaCreateAt := time.Unix(0, kcheckList.CreatedAt*int64(time.Millisecond))
				fCount = len(kcheckList.Files)
				fmt.Println("KCheck list created at:", metaCreateAt)
				for _, files := range kcheckList.Files {
					formatPath := strings.TrimPrefix(filepath.FromSlash(files.Path), string(os.PathSeparator))
					if err := checksum.CheckByHash(formatPath, files.SHA1, sha1.New()); err != nil {
						failCount++
						res <- &CheckResult{
							false,
							err,
							formatPath,
						}
						printStatus(false, formatPath)
					} else {
						passCount++
						printStatus(true, formatPath)
					}
				}
			default:
				log.Fatalf("unknown type: %s", listType)
			}
			fmt.Println("Finished.")
			fmt.Println("ALL:", fCount, "/", "PASS:", passCount, "/", "FAIL:", failCount)
			fmt.Println("Exit after 5 seconds...")
			time.Sleep(5 * time.Second)
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

func printStatus(isSuccess bool, path string) {
	successDisp := color.New(color.Bold, color.FgWhite, color.BgGreen).FprintfFunc()
	failedDisp := color.New(color.Bold, color.FgWhite, color.BgRed).FprintfFunc()
	if isSuccess {
		successDisp(color.Output, "[PASSED]")
	} else {
		failedDisp(color.Output, "[FAILED]")
	}
	fmt.Println(" ", path)
}
