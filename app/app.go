package main

import (
	"bytes"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

var globalHttpClient *http.Client

func GlobalHttpClient() *http.Client {
	if globalHttpClient == nil {
		globalHttpClient = &http.Client{}
	}

	return globalHttpClient
}

func SetGlobalHttpClient(newHttpClient *http.Client) {
	globalHttpClient = newHttpClient
}

type LoggingRoundTripper struct {
	logger    *zap.Logger
	Transport http.RoundTripper
}

func (loggingRoundTripper *LoggingRoundTripper) RoundTrip(req *http.Request) (res *http.Response, e error) {
	formattedRequestLog, err := formatRequestLog(req)
	if err != nil {
		loggingRoundTripper.logger.Error(err.Error(), zap.Error(err))
	} else {
		loggingRoundTripper.logger.Debug(formattedRequestLog)
	}

	resp, err := loggingRoundTripper.Transport.RoundTrip(req)

	if err == nil {
		formattedResponseLog, err := formatResponseLog(resp)
		if err != nil {
			loggingRoundTripper.logger.Error(err.Error())
		} else {
			loggingRoundTripper.logger.Debug(formattedResponseLog)
		}
	}

	return resp, err
}

func formatRequestLog(request *http.Request) (string, error) {
	result := strings.Builder{}

	result.Grow(1 + len(request.Method) + len(request.URL.String()) + int(request.ContentLength))
	result.WriteString(request.Method)
	result.WriteString(" ")
	result.WriteString(request.URL.String())

	for key, values := range request.Header {
		for _, value := range values {
			headerLine := fmt.Sprintf("\n%s: %s", key, value)
			result.Grow(len(headerLine))
			result.WriteString(headerLine)
		}
	}

	if request.Body != nil {
		buf := bytes.Buffer{}
		_, err := buf.ReadFrom(request.Body)
		defer request.Body.Close()
		request.Body = io.NopCloser(&buf)

		if err != nil {
			return "", err
		}

		result.Grow(buf.Len() + 1)
		result.WriteString("\n")
		result.WriteString(buf.String())
	}

	return result.String(), nil
}

func formatResponseLog(response *http.Response) (string, error) {
	result := strings.Builder{}

	result.Grow(1 + len(response.Proto) + len(response.Status))
	result.WriteString(response.Proto)
	result.WriteString(" ")
	result.WriteString(response.Status)

	for key, values := range response.Header {
		for _, value := range values {
			headerLine := fmt.Sprintf("\n%s: %s", key, value)
			result.Grow(len(headerLine))
			result.WriteString(headerLine)
		}
	}

	// todo: придумать что-то нормальное для определения файла
	if strings.Contains(response.Header.Get("Content-Type"), "image") {
		bodyHolder := "\n<file>"
		result.Grow(len(bodyHolder))
		result.WriteString(bodyHolder)
	} else {
		buf := bytes.Buffer{}
		_, err := buf.ReadFrom(response.Body)
		defer response.Body.Close()
		// overwrite
		response.Body = io.NopCloser(&buf)

		if err != nil {
			return "", err
		}

		if buf.Len() > 0 {
			result.Grow(buf.Len() + 1)
			result.WriteString("\n")
			result.WriteString(buf.String())
		}
	}

	return result.String(), nil
}
