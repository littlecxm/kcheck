package checksum

import (
	"github.com/littlecxm/kcheck/pkg/utils"
)

func Digester(done <-chan struct{}, infos <-chan utils.FileInfo, cr chan<- utils.Result) {
	for info := range infos {
		err := checkByHash(info.Path, info.FileHash, info.NewHash)
		select {
		case cr <- utils.Result{Success: err == nil, Error: err, Path: info.Path}:
		case <-done:
		}
	}
}
