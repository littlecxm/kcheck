package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/littlecxm/kcheck/kbinxml"
	"github.com/littlecxm/kcheck/kstruct"
	"golang.org/x/net/html/charset"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var builddate, commit string

const MagicNumber = 0xa042

type Option struct {
	WorkDir      string
	Listname     string
	OfficialType bool
	OfficialMeta bool
}

func main() {
	fmt.Println("kcheck v1.4.1")
	fmt.Printf("build: %s(%s)\n", builddate, commit)
	fmt.Println("--------")
	Workdir, _ := os.Getwd()
	opt := &Option{
		WorkDir:      Workdir,
		Listname:     "all.list",
		OfficialType: false,
		OfficialMeta: false,
	}

	CusType := false
	if _, err := os.Stat(opt.Listname); !os.IsNotExist(err) {
		CusType = true
	} else {
		if _, err := os.Stat("data" + string(os.PathSeparator) + "__metadata.metatxt"); !os.IsNotExist(err) {
			opt.Listname = "data" + string(os.PathSeparator) + "__metadata.metatxt"
			opt.WorkDir = Workdir + string(os.PathSeparator) + "data" + string(os.PathSeparator)
			opt.OfficialMeta = true
		} else if _, err := os.Stat("__metadata.metatxt"); !os.IsNotExist(err) {
			opt.Listname = "__metadata.metatxt"
			opt.OfficialMeta = true
		} else {
			if _, err := os.Stat("prop" + string(os.PathSeparator) + "filepath.xml"); !os.IsNotExist(err) {
				opt.Listname = "prop" + string(os.PathSeparator) + "filepath.xml"
				opt.OfficialType = true
			} else if _, err := os.Stat("filepath.xml"); !os.IsNotExist(err) {
				opt.Listname = "filepath.xml"
				opt.OfficialType = true
			} else {
				opt.OfficialType = false
			}
		}
	}

	file, err := os.Open(opt.Listname)
	if err != nil {
		log.Fatal("open file err: ", err)
	}
	defer file.Close()

	var failures []string
	var fileCount, passCount, failCount int

	if opt.OfficialType {
		// detect kbin
		isKbin := false
		magicNumber := make([]byte, 2)
		file.Read(magicNumber)
		if binary.BigEndian.Uint16(magicNumber) == MagicNumber {
			isKbin = true
		}
		file.Seek(0, io.SeekStart)
		var FilepathStruct kstruct.FilePath
		if isKbin {
			kbinByte, _ := ioutil.ReadAll(file)
			kxml, _, kbinerr := kbinxml.DeserializeKbin(kbinByte)
			if kbinerr != nil {
				log.Fatal("kbinerr", kbinerr)
			}
			err := xml.Unmarshal(kxml, &FilepathStruct)
			if err != nil {
				log.Fatal("xml.Unmarshal: ", err)
			}
		} else {
			decoder := xml.NewDecoder(file)
			decoder.CharsetReader = charset.NewReaderLabel
			err = decoder.Decode(&FilepathStruct)
			if err != nil {
				log.Fatal("xml.NewDecoder: ", err)
			}
		}
		for _, FileNode := range FilepathStruct.File {
			fileCount++
			FormatPath := strings.TrimPrefix(filepath.FromSlash(FileNode.DstPath), string(os.PathSeparator))
			if err := opt.CompareFileMD5(FormatPath, FileNode.DstMD5); err != nil {
				errstring := "[" + err.Error() + "] "
				failures = append(failures, errstring+FormatPath)
				failCount++
				fmt.Print(errstring)
			} else {
				passCount++
				fmt.Print("[PASS] ")
			}
			fmt.Println(FormatPath)
		}
	} else if opt.OfficialMeta {
		//meta
		//load metadata
		meta, err := ioutil.ReadFile(opt.Listname)
		if err != nil {
			log.Fatal(err)
		}
		//serialize metadata
		var metaStruct kstruct.MetaData
		err = json.Unmarshal(meta, &metaStruct)
		if err != nil {
			log.Fatal(err)
		}

		metaCreateAt := time.Unix(0, metaStruct.CreatedAt*int64(time.Millisecond))
		fmt.Println("metadata created at:", metaCreateAt)

		for _, files := range metaStruct.Files {
			FormatPath := strings.TrimPrefix(filepath.FromSlash(files.Path), string(os.PathSeparator))
			if err := opt.CompareFileSHA1(FormatPath, files.SHA1); err != nil {
				errstring := "[" + err.Error() + "] "
				failures = append(failures, errstring+FormatPath)
				failCount++
				fmt.Print(errstring)
			} else {
				passCount++
				fmt.Print("[PASS] ")
			}
			fmt.Println(FormatPath)
		}

	} else if CusType {
		scanner := bufio.NewScanner(file)
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 {
				continue
			}
			//TODO : bugs 因为采用空格分割，所以文件名有空格就会失败
			words := strings.Fields(scanner.Text())
			if len(words) < 1 {
				break
			}
			fileCount++
			if err := opt.CompareFileMD5(words[1], words[0]); err != nil {
				failures = append(failures, words[1])
				failCount++
				fmt.Print("[" + err.Error() + "] ")
			} else {
				passCount++
				fmt.Print("[PASS] ")
			}
			fmt.Println(words[1])
		}
	} else {
		fmt.Println("Cannot find mismatch list type")
	}

	if len(failures) > 0 {
		//写失败列表
		file2, err := os.Create("failed.list")
		if err != nil {
			log.Fatal(err)
		}

		w := bufio.NewWriter(file2)
		for _, failure := range failures {
			fmt.Fprintln(w, failure)
		}
		w.Flush()
		file2.Close()
	}

	fmt.Println("--------")
	fmt.Println("Finished.")
	fmt.Println("ALL:", fileCount, "/", "PASS:", passCount, "/", "FAIL:", failCount)
	fmt.Println("Exit after 5 seconds...")
	time.Sleep(5 * time.Second)
	os.Exit(0)
}

func (opt *Option) CompareFileMD5(relativePath, filemd5 string) error {
	fpath := opt.WorkDir + string(os.PathSeparator) + relativePath
	if _, err := os.Stat(fpath); err != nil {
		return errors.New("NOT FOUND")
	}
	// path/to/whatever exists
	f, err := os.Open(fpath)
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

func (opt *Option) CompareFileSHA1(relativePath, filesha1 string) error {
	fpath := opt.WorkDir + string(os.PathSeparator) + relativePath
	if _, err := os.Stat(fpath); err != nil {
		return errors.New("NOT FOUND")
	}
	// path/to/whatever exists
	f, err := os.Open(fpath)
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
