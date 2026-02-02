package utils

import "hash"

type FileInfo struct {
	Path     string
	FileHash string
	NewHash  func() hash.Hash
}

func SplitFiles(files []FileInfo, N int) [][]FileInfo {
	var divided [][]FileInfo
	chunkSize := (len(files) + N - 1) / N

	for i := 0; i < len(files); i += chunkSize {
		end := i + chunkSize
		if end > len(files) {
			end = len(files)
		}
		divided = append(divided, files[i:end])
	}
	return divided
}
