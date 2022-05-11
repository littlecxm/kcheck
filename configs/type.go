package configs

type MetaData struct {
	CreatedAt int64 `json:"createdAt"`
	Files     []struct {
		DPath      string `json:"dpath"`
		SPath      string `json:"spath"`
		Path       string `json:"path"`
		PathC      string `json:"pathc"`
		PathH      string `json:"pathh"`
		SHA1       string `json:"sha1"`
		DSHA1      string `json:"dsha1"`
		SSHA1      string `json:"ssha1"`
		Size       int64  `json:"size"`
		SSize      int64  `json:"ssize"`
		HashedPath string `json:"hashed_path"`
		HashedSHA1 string `json:"hashed_sha1"`
	} `json:"files"`
}

type KCheckList struct {
	CreatedAt int64         `json:"createdAt"`
	Files     []KCheckFiles `json:"files"`
}

type KCheckFiles struct {
	Path string `json:"path"`
	SHA1 string `json:"sha1"`
	Size int64  `json:"size"`
}
