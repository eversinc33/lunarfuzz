package logger

import (
	"fmt"
)

func Logln(msg string) {
	fmt.Println(msg)
}

func Log(msg string) {
	fmt.Print(msg)
}

func LogFound(path string, words string, size string) {
	fmt.Print("\033[G\033[K")
	fmt.Print(fmt.Sprintf("%s :: Words: %s, Size: %s\n", path, words, size))
}
