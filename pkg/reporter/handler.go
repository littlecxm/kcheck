package reporter

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type CheckResult struct {
	Success bool
	Error   error
	Path    string
}

func Handler(path string, res chan *CheckResult) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create result log: %s", err)
	}
	defer func() {
		_ = file.Close()
	}()

	dataWriter := bufio.NewWriter(file)
	for r := range res {
		_, _ = dataWriter.WriteString(fmt.Sprintf("[%s]: %s\n", r.Error, r.Path))
		if err := dataWriter.Flush(); err != nil {
			log.Fatal("fatal flush buf: ", err)
		}
	}
}
