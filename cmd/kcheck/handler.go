package main

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

func handler(res chan *CheckResult) {
	file, err := os.OpenFile("failed.list", os.O_APPEND|os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("failed creating result log: %s", err)
	}
	dataWriter := bufio.NewWriter(file)
	for r := range res {
		_, _ = dataWriter.WriteString(fmt.Sprintf("[%s]: %s\n", r.Error, r.Path))
		dataWriter.Flush()
	}
	file.Close()
}
