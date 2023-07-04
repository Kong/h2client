package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

func makeH2Request(url string, headerMap map[string]string, timeout int, skipVerify bool) error {
	// Create transport
	tr := &http2.Transport{}

	// Add TLS config if skipping verification
	if skipVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if strings.HasPrefix(url, "http://") {
		tr.DialTLSContext = func(ctx context.Context, netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		}
		tr.AllowHTTP = true
	}

	// Create client with timeout and transport
	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: tr,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	// Add headers to request
	for key, value := range headerMap {
		if strings.ToLower(key) == "method" {
			req.Method = value
		} else if strings.ToLower(key) == "authority" {
			req.Host = value
		} else {
			req.Header.Set(key, value)
		}
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Create map to hold headers and body
	m := make(map[string]interface{})

	// Add headers to nested map
	headers := make(map[string]interface{})
	for k, v := range resp.Header {
		if len(v) == 1 {
			headers[k] = v[0]
		} else {
			headers[k] = v
		}
	}
	headers["status"] = fmt.Sprintf("%d", resp.StatusCode)
	m["headers"] = headers

	// Add body to map
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	m["body"] = string(body)

	// Encode as JSON and print
	json, err := json.Marshal(m)
	if err != nil {
		return err
	}
	fmt.Println(string(json))

	return nil
}

func main() {
	// Note: try set GODEBUG=http2debug=1 if you are debugging this go program
	url := flag.String("url", "", "URL to make request to")
	skipVerify := flag.Bool("skip-verify", false, "Skip TLS verification")
	timeout := flag.Int("timeout", 5, "Timeout in seconds")
	headersFlag := flag.String("headers", "", "Headers to set, comma separated")
	flag.Parse()

	// Create headers map
	headerMap := make(map[string]string)
	if *headersFlag != "" {
		headersList := strings.Split(*headersFlag, ",")
		for _, header := range headersList {
			splitHeader := strings.Split(header, "=")
			key := strings.TrimPrefix(splitHeader[0], ":")
			headerMap[key] = splitHeader[1]
		}
	}

	// Make request
	err := makeH2Request(*url, headerMap, *timeout, *skipVerify)
	if err != nil {
		panic(err)
	}
}
