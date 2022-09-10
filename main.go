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
	"github.com/eversinc33/lunarfuzz/fuzz"
	"github.com/eversinc33/lunarfuzz/logger"
	"github.com/eversinc33/lunarfuzz/utils"
	color "github.com/fatih/color"
	"github.com/gernest/wow"
	"github.com/gernest/wow/spin"
	"github.com/zenthangplus/goccm"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func calibrate(browser *rod.Browser, target_url string, fast_mode bool) ([]string, []string) {
	w := wow.New(os.Stdout, spin.Get(spin.Squish), " Calibrating ...")
	w.Start()

	var target proto.TargetCreateTarget
	// Call random url that is not likely to exist to try and get a default/404 page
	target.URL = fmt.Sprintf("%s%s", target_url, utils.RandStr(10))

	page, err := browser.Page(target)

	if err != nil {
		logger.LogError(fmt.Sprintf("\nError calibrating: %s", err))
		os.Exit(1)
	}

	if !fast_mode {
		page.MustWaitLoad()
	}

	res, _ := page.HTML()

	page_words := []string{fmt.Sprint(len(strings.Split(res, " ")))}
	page_size := []string{fmt.Sprint(len(res))}

	w.PersistWith(spin.Spinner{Frames: []string{"A️"}}, fmt.Sprintf("utocalibration result: filter-size: %s, filter-words: %s", page_size[0], page_words[0]))
	return page_size, page_words
}

func doFuzz(browser *rod.Browser, target_url string, wordlist_path string, filter_size []string, filter_words []string, filter_match []string, take_screenshot bool, headers []string, fast_mode bool, max_goroutines int, output_file string) {

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

	start := time.Now()
	n_errors := 0

	result_channel := make(chan fuzz.Result)
	c := goccm.New(max_goroutines)

	for scanner.Scan() {
		fuzz_string := scanner.Text()

		go func(counter int, path string) {
			target := target_url + path
			defer c.Done()

			var r fuzz.Result
			r.IsError = false
			r.Path = target
			r.Match = false
			r.Counter = counter

			// Create new tab and navigate to page
			var p proto.TargetCreateTarget
			p.URL = ""
			page, _ := browser.Page(p)
			page.SetExtraHeaders(headers)
			err := page.Navigate(target)

			if err != nil {
				r.IsError = true
				result_channel <- r
				return
			}

			if !fast_mode {
				page.WaitLoad()
			}

			page_content, _ := page.HTML()
			r.Words = len(strings.Split(page_content, " "))
			r.Size = len(page_content)

			if filter_size != nil && !utils.Contains(filter_size, fmt.Sprint(r.Size)) {
				r.Match = true
			} else if filter_words != nil && !utils.Contains(filter_words, fmt.Sprint(r.Words)) {
				r.Match = true
			} else if filter_match != nil {
				for _, filter := range filter_match {
					if !strings.Contains(page_content, filter) {
						r.Match = true
					}
				}
			}

			if r.Match {
				if take_screenshot {
					var sc proto.PageCaptureScreenshot
					sc_bytes, _ := page.Screenshot(false, &sc)
					os.WriteFile(fmt.Sprintf("./output/%s.png", path), sc_bytes, 0644)
				}
			}

			result_channel <- r
		}(current_word, fuzz_string)

		current_word++
	}

	// TODO: verify this is the right way to handle errors
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	var f *os.File
	if output_file != "" {
		f, err = os.OpenFile(output_file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}
	}

	defer f.Close()

	for i := 1; i < current_word; i++ {
		r := <-result_channel
		if r.Match {
			logger.LogFound(r.Path, r.Words, r.Size)
			if output_file != "" {
				if _, err = f.WriteString(fmt.Sprintln(r.Path)); err != nil {
					panic(err)
				}
			}
		} else if r.IsError {
			n_errors++
		}
		logger.LogStatus(i, n_words, n_errors, r.Path)
	}

	elapsed := time.Since(start)
	logger.ClearLine()
	logger.LogPositive(fmt.Sprintf("Finished fuzzing %d urls in %s with %d errors", n_words, elapsed, n_errors))
	//c.WaitAllDone()
}

func main() {
	color.HiCyan(".-.   .-. .-..-. .-.  .--.  .----. .----..-. .-. .---. .---. ")
	color.Cyan("| |   | { } ||  `| | / {} \\ | {}  }| {_  | { } |{_   /{_   / ")
	color.Blue("| `--.| {_} || |\\  |/  /\\  \\| .-. \\| |   | {_} | /    }/    }")
	color.HiBlue("`----'`-----'`-' `-'`-'  `-'`-' `-'`-'   `-----' `---' `---'")
	fmt.Println("LunarFuzz v0.0.1")
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
	max_goroutines := flag.Int("t", 5, "Max threads. Defaults to 5")
	force_no_calibration := flag.Bool("no-ac", false, "Do not run autocalibration if no filter is given. Will output every url as a finding")
	fast_mode := flag.Bool("fast", false, "Do not wait for page to render completely")
	output_file := flag.String("o", "", "File to save all matching urls to")

	flag.Parse()

	if !strings.HasPrefix(*target_url, "http://") && !strings.HasPrefix(*target_url, "https://") {
		logger.LogAlert("Url should start with http:// or https://")
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

	if autocalibrate {
		_, filter_words = calibrate(browser, *target_url, *fast_mode) // Filtering by words is more reliable
	}
	logger.Log(fmt.Sprintf("Starting to fuzz %s with wordlist %s...", *target_url, *wordlist))
	fmt.Println()
	fmt.Println()

	doFuzz(browser, *target_url, *wordlist, filter_size, filter_words, filter_match, *take_screenshot, headers_to_use, *fast_mode, *max_goroutines, *output_file)
}
