package vrcAPI

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var DefaultHeaders = map[string]string{
	"Host":            "vrchat.com",
	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:132.0) Gecko/20100101 Firefox/132.0",
	"Accept":          "*/*",
	"Accept-Language": "en-CA,en-US;q=0.7,en;q=0.3",
	"Referer":         "https://vrchat.com/home?utm-source=hello-login",
	"DNT":             "1",
	"Priority":        "u=4",
}

const (
	API_ENDPOINT = "https://vrchat.com/api/1"
	MAX_TIMEOUT  = time.Duration(time.Second * 15)
)

type Client struct {
	Config *Configuration
	Client *http.Client
}

func (c *Client) NewRequest(method, url string, body io.Reader) (*http.Request, error) {
	cctx, _ := context.WithTimeout(context.Background(), MAX_TIMEOUT)
	req, err := http.NewRequestWithContext(cctx, method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range DefaultHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Cookie", fmt.Sprintf("auth=%s; twoFactorAuth=%s", c.Config.AuthCookie, c.Config.TwoFactorAuth))
	return req, nil
}
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		err = fmt.Errorf("Client.Do - c.Client.Do: %w", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := decompressBody(resp)

	if err != nil {
		err = fmt.Errorf("Client.Do - decompressBody: %w", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("%s HTTP %s: %s", req.URL.String(), resp.Status, body)
		return nil, err
	}

	resp.Body = io.NopCloser(strings.NewReader(body))
	return resp, nil
}

func (c *Client) DoWDefaults(req *http.Request) (*http.Response, error) {
	for k, v := range DefaultHeaders {
		req.Header.Set(k, v)
	}
	return c.Do(req)
}

func decompressBody(resp *http.Response) (string, error) {
	var reader io.ReadCloser
	var err error

	// Check the Content-Encoding header to determine how to decompress
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return "", err
		}
		defer reader.Close()
	case "deflate":
		reader = flate.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
		defer reader.Close()
	}

	// Read the decompressed data
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		return "", err
	}

	return buf.String(), nil
}
