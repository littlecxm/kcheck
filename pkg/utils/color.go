package utils

import (
	"fmt"
	"github.com/fatih/color"
)

func PrintStatus(isSuccess bool, path string) {
	successDisp := color.New(color.Bold, color.FgWhite, color.BgGreen).FprintfFunc()
	failedDisp := color.New(color.Bold, color.FgWhite, color.BgRed).FprintfFunc()
	if isSuccess {
		successDisp(color.Output, "[PASSED]")
	} else {
		failedDisp(color.Output, "[FAILED]")
	}
	fmt.Println(" ", path)
}
