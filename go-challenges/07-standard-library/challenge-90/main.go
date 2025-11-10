package main

import (
	"net/url"
)

func ParseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

func GetHost(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}

func AddQueryParam(rawURL, key, value string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	
	q := u.Query()
	q.Add(key, value)
	u.RawQuery = q.Encode()
	
	return u.String(), nil
}

func main() {}
