package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/eversinc33/lunarfuzz/driver"
	"github.com/eversinc33/lunarfuzz/logger"
	"github.com/eversinc33/lunarfuzz/utils"
	color "github.com/fatih/color"
	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func calibrate(browser *rod.Browser, target_url string) ([]string, []string) {
	w := wow.New(os.Stdout, spin.Get(spin.Squish), " Calibrating ...")
	w.Start()

	var target proto.TargetCreateTarget
	// Call random url that is not likely to exist to try and get a default/404 page
	target.URL = fmt.Sprintf("%s%s", target_url, utils.RandStr(10))

	bogus_response, err := browser.Page(target)
	if err != nil {
		logger.LogError(fmt.Sprintf("\nError calibrating: %s", err))
		os.Exit(1)
	}
	res, _ := bogus_response.HTML()

	page_words := []string{fmt.Sprint(len(strings.Split(res, " ")))}
	page_size := []string{fmt.Sprint(len(res))}

	w.PersistWith(spin.Spinner{Frames: []string{"A️"}}, fmt.Sprintf("utocalibration found size: %s, words: %s", page_size[0], page_words[0]))
	return page_size, page_words
}

func fuzz(browser *rod.Browser, target_url string, wordlist_path string, filter_size []string, filter_words []string, filter_match []string, take_screenshot bool, autocalibrate bool, headers []string, fast_mode bool) {

	wordlist, err := os.Open(wordlist_path)
	if err != nil {
		log.Fatal(err)
	}
	defer wordlist.Close()

	n_words, err := utils.CountLines(wordlist)
	current_word := 1

	if err != nil {
		log.Fatal(err)
	}

	_, err = wordlist.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(wordlist)

	if autocalibrate {
		_, filter_words = calibrate(browser, target_url) // Filtering by words is more reliable
	}

	start := time.Now()
	n_errors := 0

	for scanner.Scan() {
		fuzz := scanner.Text()
		target := target_url + fuzz

		page := browser.MustPage("")
		page.SetExtraHeaders(headers)
		err := page.Navigate(target)

		if err != nil {
			n_errors++
			fmt.Print("\033[G\033[K")
			logger.Log(fmt.Sprintf("[%d/%d] Errors: %d :: %s", current_word, n_words, n_errors, target))
			current_word++
			continue
		}

		if !fast_mode {
			page.MustWaitLoad()
		}
		page_content, _ := page.HTML()
		page_words := fmt.Sprint(len(strings.Split(page_content, " ")))
		page_size := fmt.Sprint(len(page_content))

		found := false
		if filter_size != nil && !utils.Contains(filter_size, page_size) {
			found = true
		} else if filter_words != nil && !utils.Contains(filter_words, page_words) {
			found = true
		} else if filter_match != nil {
			for _, filter := range filter_match {
				if !strings.Contains(page_content, filter) {
					found = true
				}
			}
		}

		if found {
			logger.LogFound(target, page_words, page_size)

			if take_screenshot {
				page.MustScreenshot(fmt.Sprintf("output/%s.png", fuzz))
			}
		}

		fmt.Print("\033[G\033[K")
		logger.Log(fmt.Sprintf("[%d/%d] Errors: %d :: %s", current_word, n_words, n_errors, target))
		current_word++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	elapsed := time.Since(start)
	fmt.Println()
	logger.LogResult(fmt.Sprintf("Finished fuzzing %d urls in %s", n_words, elapsed))
}

func main() {
	color.HiCyan(".-.   .-. .-..-. .-.  .--.  .----. .----..-. .-. .---. .---. ")
	color.Cyan("| |   | { } ||  `| | / {} \\ | {}  }| {_  | { } |{_   /{_   / ")
	color.Blue("| `--.| {_} || |\\  |/  /\\  \\| .-. \\| |   | {_} | /    }/    }")
	color.HiBlue("`----'`-----'`-' `-'`-'  `-'`-' `-'`-'   `-----' `---' `---'")
	color.Yellow("LunarFuzz v0.0.1")
	fmt.Println()

	// TODO: use better flag library
	target_url := flag.String("u", "", "Target url")
	wordlist := flag.String("w", "", "Wordlist to use")
	fs := flag.String("fs", "", "Filter response by size")
	fw := flag.String("fw", "", "Filter response by words")
	fm := flag.String("fm", "", "Filter response by string match")
	cookies := flag.String("b", "", "Cookies to use")
	headers := flag.String("H", "", "Headers to use in the format of 'Header: Value; Header: Value'")
	take_screenshot := flag.Bool("screenshot", false, "Save screenshots for matches")
	force_no_calibration := flag.Bool("no-ac", false, "Do not run autocalibration if no filter is given. Will output every url as a finding")
	fast_mode := flag.Bool("fast", false, "Do not wait for page to render completely")

	flag.Parse()

	if !strings.HasPrefix(*target_url, "http://") && !strings.HasPrefix(*target_url, "https://") {
		logger.LogResult("Url should start with http:// or https://")
		os.Exit(1)
	}
	if !strings.HasSuffix(*target_url, "/") {
		*target_url += "/"
	}

	filter_size := utils.SplitOrNil(*fs, ",")
	filter_words := utils.SplitOrNil(*fw, ",")
	filter_match := utils.SplitOrNil(*fm, ",")

	autocalibrate := false
	if !*force_no_calibration {
		if filter_size == nil && filter_words == nil && filter_match == nil {
			autocalibrate = true
		}
	}

	if *take_screenshot {
		newpath := filepath.Join(".", "output")
		err := os.MkdirAll(newpath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	headers_to_use := driver.ParseHeaders(headers)

	browser := driver.SetupBrowser(cookies)
	defer browser.MustClose()

	fuzz(browser, *target_url, *wordlist, filter_size, filter_words, filter_match, *take_screenshot, autocalibrate, headers_to_use, *fast_mode)
}
