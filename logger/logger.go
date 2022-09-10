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

func LogFound(path string, words int, size int) {
	fmt.Print("\033[G\033[K")
	LogPositive(fmt.Sprintf("%s :: Words: %d, Size: %d\n", path, words, size))
}

func LogAlert(msg string) {
	color.Yellow(fmt.Sprintf("%s", msg))
}

func LogPositive(msg string) {
	color.Green(fmt.Sprintf("%s", msg))
}

func ClearLine() {
	fmt.Print("\033[G\033[K")
}

func LogStatus(current_word int, n_words int, n_errors int, target string) {
	ClearLine()
	Log(fmt.Sprintf("[%d/%d] Errors: %d :: %s", current_word, n_words, n_errors, target))
}
