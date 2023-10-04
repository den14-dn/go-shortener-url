package shortener

import "github.com/speps/go-hashids/v2"

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
