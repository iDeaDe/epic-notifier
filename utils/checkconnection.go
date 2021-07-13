package utils

import (
	"net/http"
	"time"
)

func CheckConnection(url string) error {
	conn := http.Client{
		Timeout: time.Second * 4,
	}

	resp, err := conn.Get(url)
	if resp != nil &&
		err == nil &&
		int(resp.StatusCode/100) != 5 {
		return nil
	}

	return err
}
