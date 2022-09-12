package driver

import (
	"strings"
    "time"

	"github.com/go-rod/rod/lib/proto"
)

func ParseCookies(cookies *string, domain string) []proto.NetworkCookie {
	cookies_to_use := []proto.NetworkCookie{}

	if *cookies == "" {
		return cookies_to_use
	}

	for _, c := range strings.Split(*cookies, "; ") {
		ck := strings.Split(c, "=")
		cookies_to_use = append(cookies_to_use, proto.NetworkCookie{
			Name:  ck[0],
			Value: ck[1],
            Domain: domain,
            Path: "/",
            Expires: proto.TimeSinceEpoch(time.Now().Add(time.Hour * 12).Unix()),
		})
	}

	return cookies_to_use
}
