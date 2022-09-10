package logger

import (
	"fmt"

	"github.com/fatih/color"
)

func Logln(msg string) {
	fmt.Println(msg)
}

func Log(msg string) {
	fmt.Print(msg)
}

func LogError(msg string) {
	color.Red(msg)
}

func LogFound(path string, words string, size string) {
	fmt.Print("\033[G\033[K")
	color.Green(fmt.Sprintf("%s :: Words: %s, Size: %s\n", path, words, size))
}

func LogResult(msg string) {
	color.Yellow(fmt.Sprintf("%s", msg))
}
