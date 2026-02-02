package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/beevik/etree"
	"github.com/fatih/color"
	"github.com/littlecxm/kbinxml-go"
	"github.com/littlecxm/kcheck/configs"
	"github.com/littlecxm/kcheck/pkg/checksum"
	"github.com/littlecxm/kcheck/pkg/format"
	"github.com/littlecxm/kcheck/pkg/reporter"
	"github.com/littlecxm/kcheck/pkg/utils"
	"github.com/urfave/cli/v2"
)

func commandHandler(c *cli.Context) error {
	fmt.Sprintf("kcheck %s %s (built: %s)", version, commitID, buildDate)

	var listPath string
	if c.NArg() == 0 {
		guessListPath, err := format.GuessListPath()
		if err != nil {
			fmt.Println("use", color.GreenString("help"), "to get more info")
			os.Exit(-1)
		}
		listPath = guessListPath
		log.Println("use default file:", listPath)
	}

	if c.NArg() > 0 {
		manualPath := c.Args().Get(0)
		listPath = filepath.Base(manualPath)
		configs.WorkDir = filepath.Dir(manualPath)
	}
	log.Println("current work dir:", configs.WorkDir)
	log.Println("list path:", listPath)

	f, err := os.Open(filepath.Join(configs.WorkDir, listPath))
	if err != nil {
		log.Fatalf("faild open file: %s", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if listType == "" {
		listType, err = format.GuessType(listPath)
		if err != nil {
			log.Fatalf("get list type error: %s", err)
		}
		if listType == configs.KCheckType {
			log.Println("unknown file type, use default:", configs.KCheckType)
		}
	}

	inByte, err := io.ReadAll(f)
	if err != nil {
		log.Fatalf("load file error: %s", err)
	}
	var fileInfos []utils.FileInfo
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
		hashFactory := md5.New
		for _, fNode := range listNode.SelectElements("file") {
			dstPath := fNode.SelectElement("dst_path").Text()
			dstMd5 := fNode.SelectElement("dst_md5").Text()
			formatPath := strings.TrimPrefix(filepath.FromSlash(dstPath), string(os.PathSeparator))
			fileInfos = append(fileInfos, utils.FileInfo{Path: formatPath, FileHash: dstMd5, NewHash: hashFactory})
		}

	case configs.MetadataType:
		var metaStruct configs.MetaData

		if err := json.Unmarshal(inByte, &metaStruct); err != nil {
			log.Fatal("load metadata list err:", err)
		}

		metaCreateAt := time.Unix(0, metaStruct.CreatedAt*int64(time.Millisecond))
		fCount = len(metaStruct.Files)
		hashFactory := sha1.New
		log.Println("metadata created at:", metaCreateAt)
		for _, files := range metaStruct.Files {
			fileSHA1 := files.SHA1
			if fileSHA1 == "" {
				fileSHA1 = files.SSHA1
			}

			filePath := files.Path
			if filePath == "" {
				filePath = files.SPath
			}

			formatPath := filepath.Join(
				"data",
				strings.TrimPrefix(filepath.FromSlash(filePath), string(os.PathSeparator)),
			)
			fileInfos = append(fileInfos, utils.FileInfo{Path: formatPath, FileHash: fileSHA1, NewHash: hashFactory})
		}

	case configs.KCheckType:
		var kcheckList configs.KCheckList
		if err := json.Unmarshal(inByte, &kcheckList); err != nil {
			log.Fatal("load KCheck list err:", err)
		}
		metaCreateAt := time.Unix(0, kcheckList.CreatedAt*int64(time.Millisecond))
		fCount = len(kcheckList.Files)
		hashFactory := sha256.New
		fmt.Println("KCheck list created at:", metaCreateAt)
		for _, files := range kcheckList.Files {
			formatPath := filepath.Join(
				configs.WorkDir,
				strings.TrimPrefix(filepath.FromSlash(files.Path), string(os.PathSeparator)),
			)
			fileInfos = append(fileInfos, utils.FileInfo{Path: formatPath, FileHash: files.SHA256, NewHash: hashFactory})
		}

	default:
		log.Fatalf("unknown type: %s", listType)
	}

	// report handler
	reporter.Init("failed.list")
	res := make(chan utils.Result)

	// done
	done := make(chan struct{})
	defer close(done)

	// fan in file paths
	paths := make(chan utils.FileInfo)
	go func() {
		for _, info := range fileInfos {
			paths <- info
		}
		close(paths)
	}()

	// consumer
	var wg sync.WaitGroup
	const numDigesters = 5
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			checksum.Digester(done, paths, res)
			wg.Done()
		}()
	}
	// solve result err
	go func() {
		wg.Wait()
		close(res)
	}()

	for r := range res {
		if r.Error != nil {
			failCount++
			utils.PrintStatus(false, r.Path)
			reporter.WriteResult(r)
		} else {
			passCount++
			utils.PrintStatus(true, r.Path)
		}
	}
	reporter.Close()

	fmt.Printf(
		"Finished.\n"+
			"ALL: %d / PASS: %d / FAIL: %d\n"+
			"Exit after 5 seconds...",
		fCount,
		passCount,
		failCount,
	)
	time.Sleep(5 * time.Second)
	return nil
}
