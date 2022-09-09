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

	"github.com/eversinc33/lunarfuzz/logger"
	"github.com/eversinc33/lunarfuzz/utils"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func calibrate(browser *rod.Browser, target_url string) ([]string, []string) {
	logger.Logln("Calibrating...")
	bogus_response, err := browser.MustPage(fmt.Sprintf("%saeiavnevnhafviauhoe", target_url)).HTML() // TODO: use real randomness
	if err != nil {
		log.Fatal(fmt.Sprintf("Error calibrating: %s", err))
	}
	page_words := []string{fmt.Sprint(len(strings.Split(bogus_response, " ")))}
	page_size := []string{fmt.Sprint(len(bogus_response))}
	logger.Log(fmt.Sprintf("Found size: %s, words: %s\n", page_size[0], page_words[0]))
	return page_size, page_words
}

func fuzz(target_url string, wordlist_path string, filter_size []string, filter_words []string, cookies_to_use []proto.NetworkCookie, take_screenshot bool, autocalibrate bool) {
	browser := rod.New().MustConnect()
	defer browser.MustClose()

	for _, cookie := range cookies_to_use {
		browser.MustSetCookies(&cookie)
	}

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

	for scanner.Scan() {
		word := scanner.Text()
		target := target_url + word

		page := browser.MustPage(target)
		page_content, _ := page.HTML()
		page_words := fmt.Sprint(len(strings.Split(page_content, " ")))
		page_size := fmt.Sprint(len(page_content))

		found := false
		if filter_size != nil && !utils.Contains(filter_size, page_size) {
			found = true
		} else if filter_words != nil && !utils.Contains(filter_words, page_words) {
			found = true
		}

		if found {
			logger.LogFound(target, page_words, page_size)
		}

		if found && take_screenshot {
			page.MustScreenshot(fmt.Sprintf("output/%s.png", word))
		}

		fmt.Print("\033[G\033[K")
		logger.Log(fmt.Sprintf(":: [%d/%d] :: %s", current_word, n_words, target))

		current_word++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	fmt.Println("Lunar v0.0.1")

	target_url := flag.String("u", "", "Target url")
	wordlist := flag.String("w", "", "Wordlist to use")
	fs := flag.String("fs", "", "Filter response by size")
	fw := flag.String("fw", "", "Filter response by words")
	cookies := flag.String("b", "", "Cookies to use")
	take_screenshot := flag.Bool("screenshot", false, "Save screenshots for matches")

	flag.Parse()

	if !strings.HasPrefix(*target_url, "http://") && !strings.HasPrefix(*target_url, "https://") {
		fmt.Println("[!] Url should start with http:// or https://")
		os.Exit(1)
	}
	if !strings.HasSuffix(*target_url, "/") {
		*target_url += "/"
	}

	var filter_size []string
	var filter_words []string

	if *fs == "" {
		filter_size = nil
	} else {
		filter_size = strings.Split(*fs, ",")
	}

	if *fw == "" {
		filter_words = nil
	} else {
		filter_words = strings.Split(*fw, ",")
	}

	autocalibrate := false
	if filter_words == nil && filter_size == nil {
		autocalibrate = true
	}

	if *take_screenshot {
		newpath := filepath.Join(".", "output")
		err := os.MkdirAll(newpath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	cookies_to_use := []proto.NetworkCookie{}
	if *cookies != "" {
		for _, c := range strings.Split(*cookies, "; ") {
			ck := strings.Split(c, "=")
			cookies_to_use = append(cookies_to_use, proto.NetworkCookie{
				Name:  ck[0],
				Value: ck[1],
			})
		}

	}
	fuzz(*target_url, *wordlist, filter_size, filter_words, cookies_to_use, *take_screenshot, autocalibrate)
}
