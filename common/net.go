package common

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type StatusCodeError struct {
	StatusCode int
	URL        string
}

func (e *StatusCodeError) Error() string {
	return fmt.Sprintf(
		"client failure: HTTP %v %s for url: %s",
		e.StatusCode,
		http.StatusText(e.StatusCode),
		e.URL,
	)
}

func CheckStatusCode(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		return &StatusCodeError{resp.StatusCode, resp.Request.URL.String()}
	}
	return nil
}

func DownloadFile(client *http.Client, url string, path string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	if err := CheckStatusCode(resp); err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create %q: %v", path, err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to %q: %v", path, err)
	}
	return nil
}
