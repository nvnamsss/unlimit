package utility

import (
	"encoding/base64"
	"io"
	"net/http"
)

func Base64ToImageData(b64 string) ([]byte, error) {
	// Dummy implementation for illustration purposes
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FetchFromUrl fetches data from a given URL
func FetchFromUrl(url string) []byte {
	imageResp, _ := http.Get(url)

	imageBytes, _ := io.ReadAll(imageResp.Body)
	return imageBytes
}
