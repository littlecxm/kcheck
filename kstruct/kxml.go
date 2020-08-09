package kstruct

import "encoding/xml"

type FilePath struct {
	XMLName xml.Name   `xml:"list"`
	NrFiles string     `xml:"nr_files"`
	SzTotal string     `xml:"sz_total"`
	File    []FileNode `xml:"file"`
}

type FileNode struct {
	XMLName xml.Name `xml:"file"`
	Name    string   `xml:"name,attr"`
	DstPath string   `xml:"dst_path"`
	DstMD5  string   `xml:"dst_md5"`
	DstSize string   `xml:"dst_size"`
}
