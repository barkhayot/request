package request

import (
	"fmt"
	"net/url"
)

func validateProxy(proxyStr string) (*url.URL, error) {
	if proxyStr == "" {
		return nil, fmt.Errorf("proxy string is empty")
	}

	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	return proxyURL, nil
}
