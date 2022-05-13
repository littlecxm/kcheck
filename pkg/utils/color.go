package utils

import (
	"fmt"
	"github.com/fatih/color"
)

var (
	successDisp = color.New(color.Bold, color.FgWhite, color.BgGreen).FprintfFunc()
	failedDisp  = color.New(color.Bold, color.FgWhite, color.BgRed).FprintfFunc()
)

func PrintStatus(isSuccess bool, path string) {
	if isSuccess {
		successDisp(color.Output, "[PASSED]")
	} else {
		failedDisp(color.Output, "[FAILED]")
	}
	fmt.Println(" ", path)
}
