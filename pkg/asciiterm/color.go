package asciiterm

import "fmt"

const (
	eraseCurrentTerminalLine = "\033[2K\r"
	checkBox                 = "\u2705"
	infoColor                = checkBox + " \033[1;34m"
	warningColor             = "\033[1;33m"
	asciiColorEnd            = "\033[0m"
)

func PrintfWarn(format string, msg ...interface{}) {
	fmt.Printf(eraseCurrentTerminalLine+warningColor+format+asciiColorEnd, msg...)
}

func PrintfInfo(format string, msg ...interface{}) {
	fmt.Printf(eraseCurrentTerminalLine+infoColor+format+asciiColorEnd, msg...)
}
