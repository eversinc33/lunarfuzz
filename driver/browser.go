package driver

import (
  "github.com/go-rod/rod"
)

func SetupBrowser(cookies *string, domain string) *rod.Browser {
	browser := rod.New().MustConnect()

	cookies_to_use := ParseCookies(cookies, domain)
	for _, cookie := range cookies_to_use {
		browser.MustSetCookies(&cookie)
	}

	return browser
}
