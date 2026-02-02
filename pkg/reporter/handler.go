package reporter

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/littlecxm/kcheck/pkg/utils"
)

var file *os.File
var dataWriter *bufio.Writer

func Init(path string) {
	var err error
	file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create result log: %s", err)
	}
	dataWriter = bufio.NewWriter(file)
}

// WriteResult appends a result entry to the report file.
func WriteResult(r utils.Result) {
	if dataWriter == nil {
		return
	}
	_, _ = dataWriter.WriteString(fmt.Sprintf("[%v]: %s\n", r.Error, r.Path))
	if err := dataWriter.Flush(); err != nil {
		log.Fatal("fatal flush buf: ", err)
	}
}

// Close flushes and closes the report file.
func Close() {
	if dataWriter != nil {
		_ = dataWriter.Flush()
	}
	if file != nil {
		_ = file.Close()
	}
}
