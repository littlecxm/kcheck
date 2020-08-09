package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
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

func main() {
	fmt.Println("kcheck v1.3")
	fmt.Println("--------")
	//Workdir,_ := os.Getwd()
	//Listname := Workdir + "all.list"
	Listname := "all.list"
	CusType := false
	OfficialType := false
	OfficialMeta := false

	if _, err := os.Stat(Listname); !os.IsNotExist(err) {
		CusType = true
	} else {
		Listname = "filepath.xml"
		if _, err := os.Stat(Listname); !os.IsNotExist(err) {
			fStream, _ := ioutil.ReadFile(Listname)
			if strings.Index(string(fStream), "<list>") == -1 {
				//not detect file
				log.Fatal("filepath.xml load fail")
			}
			OfficialType = true
		} else {
			Listname = "__metadata.metatxt"
			if _, err := os.Stat(Listname); !os.IsNotExist(err) {
				OfficialMeta = true
			} else {
				OfficialType = false
			}
		}
	}

	file, err := os.Open(Listname)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var failures []string
	var fileCount, passCount, failCount int

	if OfficialType {
		var FilepathStruct kstruct.FilePath
		decoder := xml.NewDecoder(file)
		decoder.CharsetReader = charset.NewReaderLabel
		err = decoder.Decode(&FilepathStruct)
		if err != nil {
			log.Fatal(err)
		}
		for _, FileNode := range FilepathStruct.File {
			fileCount++
			FormatPath := strings.TrimPrefix(filepath.FromSlash(FileNode.DstPath), string(os.PathSeparator))
			if err := CompareFileMD5(FormatPath, FileNode.DstMD5); err != nil {
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
	} else if OfficialMeta {
		//meta
		//load metadata
		meta, err := ioutil.ReadFile(Listname)
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
			if err := CompareFileSHA1(FormatPath, files.SHA1); err != nil {
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
			if err := CompareFileMD5(words[1], words[0]); err != nil {
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

func CompareFileMD5(relativePath, filemd5 string) error {
	WorkDir, _ := os.Getwd()
	fpath := WorkDir + string(os.PathSeparator) + relativePath
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

func CompareFileSHA1(relativePath, filesha1 string) error {
	WorkDir, _ := os.Getwd()
	fpath := WorkDir + string(os.PathSeparator) + relativePath
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
