package main

import (
	"io"
	"os"
	"path/filepath"
)

func getWorkdir() string {
	resultWorkdir := os.Getenv("WORKDIR")

	if resultWorkdir == "" {
		var err error
		executable, err := os.Executable()
		if err != nil {
			resultWorkdir = "."
		} else {
			resultWorkdir = filepath.Dir(executable)
		}
	}

	return filepath.Clean(resultWorkdir)
}

func downloadFileByLink(link string) (io.ReadCloser, error) {
	resp, err := GlobalHttpClient().Get(link)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}