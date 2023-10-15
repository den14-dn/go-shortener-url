// Package shortener using a third party library hashes a string value.
package shortener

import "github.com/speps/go-hashids/v2"

// ShortenURL uses a third party library github.com/speps/go-hashids/v2' to shorten URLs.
func ShortenURL(origURL string) (string, error) {
	hid := hashids.NewData()
	hid.Salt = origURL
	hi, err := hashids.NewWithData(hid)
	if err != nil {
		return "", err
	}
	id, err := hi.Encode([]int{45, 434, 1313, 99})
	if err != nil {
		return "", err
	}
	return id, nil
}
