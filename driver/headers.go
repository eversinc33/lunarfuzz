package driver

import (
	"strings"
)

func ParseHeaders(headers *string) []string {
	headers_to_use := []string{}

	if *headers == "" {
		return headers_to_use
	}

	for _, h := range strings.Split(*headers, "; ") {
		header := strings.Split(h, ": ")
		// rod does not have a dedicated header struct, so it uses this weird way of a string slice
		headers_to_use = append(headers_to_use, header[0])
		headers_to_use = append(headers_to_use, header[1])
	}

	return headers_to_use
}
