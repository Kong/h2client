package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

func makeH2Request(
	method string, url string,
	headerMap map[string]string, requestBody io.Reader,
	timeout int, tr http.RoundTripper, stremMode bool) error {

	// Create client with timeout and transport
	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: tr,
	}

	var reqBody *bytes.Buffer

	if requestBody == nil || stremMode {
		reqBody = new(bytes.Buffer)
	} else {
		// convert the buffer to a interface that supports `.Len()` so that Content-Length header is added
		b, err := io.ReadAll(requestBody)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return err
	}

	// Add headers to request
	for key, value := range headerMap {
		if strings.ToLower(key) == "method" {
			req.Method = value
		} else if strings.ToLower(key) == "authority" {
			req.Host = value
		} else if strings.ToLower(key) == "path" {
			req.URL.Path = value
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
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	m["body"] = string(responseBody)

	// Encode as JSON and print
	jsonResponse, err := json.Marshal(m)
	if err != nil {
		return err
	}
	fmt.Println(string(jsonResponse))

	return nil
}

func makeHttp2Transport(url string, tlsClientConfig *tls.Config) http.RoundTripper {
	tr := &http2.Transport{TLSClientConfig: tlsClientConfig}

	if strings.HasPrefix(url, "http://") {
		tr.DialTLSContext = func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		}
		tr.AllowHTTP = true
	}

	return tr
}

func main() {
	// Note: try set GODEBUG=http2debug=1 if you are debugging this go program
	url := flag.String("url", "", "URL to make request to")
	skipVerify := flag.Bool("skip-verify", false, "Skip TLS verification")
	timeout := flag.Int("timeout", 5, "Timeout in seconds")
	headersFlag := flag.String("headers", "", "Headers to set, comma separated")
	http1Flag := flag.Bool("http1", false, "Use HTTP/1.[01] protocol")
	postFlag := flag.Bool("post", false, "Use POST, body is read from standard input")
	streamMode := flag.Bool("stream", false, "Use stream mode in HTTP2 transport, which means the request has no 'Content-Length' header")
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

	var tlsClientConfig tls.Config

	// Add TLS config if skipping verification
	if *skipVerify {
		tlsClientConfig.InsecureSkipVerify = true
	}

	var tr http.RoundTripper
	if *http1Flag {
		tr = &http.Transport{TLSClientConfig: &tlsClientConfig}
	} else {
		tr = makeHttp2Transport(*url, &tlsClientConfig)
	}

	var method string
	var body io.Reader
	if *postFlag {
		method = "POST"
		body = os.Stdin
	} else {
		method = "GET"
	}

	// Make request
	err := makeH2Request(method, *url, headerMap, body, *timeout, tr, *streamMode)
	if err != nil {
		panic(err)
	}
}
