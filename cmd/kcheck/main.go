package main

import (
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
				listPath = guessPath()
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
				listPath = c.Args().Get(0)
			}

			f, err := os.Open(listPath)
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
			successDisp := color.New(color.Bold, color.FgWhite, color.BgGreen).FprintfFunc()
			failedDisp := color.New(color.Bold, color.FgWhite, color.BgRed).FprintfFunc()
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
					if err := checksum.CompareFileMD5(formatPath, dstMd5); err != nil {
						failCount++
						res <- &CheckResult{
							false,
							err,
							formatPath,
						}
						failedDisp(color.Output, "[FAILED]")
					} else {
						passCount++
						successDisp(color.Output, "[PASSED]")
					}
					fmt.Println(" ", formatPath)
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
					if err := checksum.CompareFileSHA1(formatPath, fileSHA1); err != nil {
						failCount++
						res <- &CheckResult{
							false,
							err,
							formatPath,
						}
						failedDisp(color.Output, "[FAILED]")
					} else {
						passCount++
						successDisp(color.Output, "[PASSED]")
					}
					fmt.Println(" ", formatPath)
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
					if err := checksum.CompareFileSHA1(formatPath, files.SHA1); err != nil {
						failCount++
						res <- &CheckResult{
							false,
							err,
							formatPath,
						}
						failedDisp(color.Output, "[FAILED]")
					} else {
						passCount++
						successDisp(color.Output, "[PASSED]")
					}
					fmt.Println(" ", formatPath)
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
