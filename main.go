package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode/utf8"

	http "github.com/bogdanfinn/fhttp"
	tls_client_cffi "github.com/bogdanfinn/tls-client/cffi_src"
	"github.com/cristalhq/base64"
	"github.com/google/uuid"
	proxy "github.com/tomtom103/cffi-proxy/goproxy"
)

var port int

func init() {
	flag.IntVar(&port, "port", 8000, "Port where the proxy will listen on")
}

func destroyAll() {
	tls_client_cffi.ClearSessionCache()
}

func destroySession(sessionId string) {
	tls_client_cffi.RemoveSession(sessionId)
}

// Response represents the structure of the response sent back to the client.
type Response struct {
	Id           string              `json:"id"`
	Body         string              `json:"body"`
	Cookies      map[string]string   `json:"cookies"`
	Headers      map[string][]string `json:"headers"`
	SessionId    string              `json:"sessionId,omitempty"`
	Status       int                 `json:"status"`
	Target       string              `json:"target"`
	UsedProtocol string              `json:"usedProtocol"`
	IsBase64     bool                `json:"isBase64,omitempty"`
}

// ExtendedRequestInput extends the RequestInput with additional fields.
type ExtendedRequestInput struct {
	tls_client_cffi.RequestInput
	WantHistory    bool `json:"wantHistory"`
	DetectEncoding bool `json:"detectEncoding"`
}

// BuildResponse constructs the Response object from the HTTP response.
func BuildResponse(
	sessionId string,
	withSession bool,
	resp *http.Response,
	cookies []*http.Cookie,
	detect bool,
) (Response, *tls_client_cffi.TLSClientError) {
	defer resp.Body.Close()

	ce := resp.Header.Get("Content-Encoding")

	var respBodyBytes []byte
	var err error

	if !resp.Uncompressed {
		resp.Body = http.DecompressBodyByType(resp.Body, ce)
	}

	respBodyBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		clientErr := tls_client_cffi.NewTLSClientError(err)
		return Response{}, clientErr
	}

	var finalResponse string

	isBase64 := detect && !utf8.Valid(respBodyBytes)
	if isBase64 {
		finalResponse = base64.StdEncoding.EncodeToString(respBodyBytes)
	} else {
		finalResponse = string(respBodyBytes)
	}

	response := Response{
		Id:           uuid.New().String(),
		Status:       resp.StatusCode,
		UsedProtocol: resp.Proto,
		Body:         finalResponse,
		Headers:      resp.Header,
		Target:       "",
		Cookies:      cookiesToMap(cookies),
		IsBase64:     isBase64,
	}

	if resp.Request != nil && resp.Request.URL != nil {
		response.Target = resp.Request.URL.String()
	}

	if withSession {
		response.SessionId = sessionId
	}

	return response, nil
}

func cookiesToMap(cookies []*http.Cookie) map[string]string {
	ret := make(map[string]string, 0)

	for _, c := range cookies {
		ret[c.Name] = c.Value
	}

	return ret
}

func handleErrorResponse(sessionId string, withSession bool, err *tls_client_cffi.TLSClientError) *Response {
	response := Response{
		Id:      uuid.New().String(),
		Status:  0,
		Body:    err.Error(),
		Headers: nil,
		Cookies: nil,
	}

	if withSession {
		response.SessionId = sessionId
	}

	return &response
}

func buildCookies(cookies []tls_client_cffi.Cookie) []*http.Cookie {
	var ret []*http.Cookie

	for _, cookie := range cookies {
		ret = append(ret, &http.Cookie{
			Name:    cookie.Name,
			Value:   cookie.Value,
			Path:    cookie.Path,
			Domain:  cookie.Domain,
			Expires: cookie.Expires.Time,
		})
	}

	return ret
}

// handleRequest processes incoming HTTP requests using tls-client.
func handleRequest(logger *log.Logger) func(req *http.Request, ctx *proxy.ProxyCtx) (*http.Request, *http.Response) {
	return func(req *http.Request, ctx *proxy.ProxyCtx) (*http.Request, *http.Response) {
		logger.Printf("Handling request for %s", req.URL)

		// Convert the incoming *http.Request to ExtendedRequestInput
		requestInput := ExtendedRequestInput{
			RequestInput: tls_client_cffi.RequestInput{
				RequestMethod: req.Method,
				RequestUrl:    req.URL.String(),
				Headers:       make(map[string]string),
			},
			DetectEncoding: true,
		}

		// Copy headers from the incoming request
		for name, values := range req.Header {
			requestInput.RequestInput.Headers[name] = values[0]
		}

		requestInput.RequestInput.Headers["X-Hello-World"] = "true"

		// Handle the request using tls-client
		response := request(&requestInput)

		// Build the *http.Response to return to the client
		resp := &http.Response{
			StatusCode: response.Status,
			Header:     response.Headers,
			Body:       io.NopCloser(strings.NewReader(response.Body)),
			Request:    req,
		}

		return req, resp
	}
}

func request(requestInput *ExtendedRequestInput) *Response {
	tlsClient, sessionId, withSession, err := tls_client_cffi.CreateClient(requestInput.RequestInput)
	if err != nil {
		return handleErrorResponse(sessionId, withSession, err)
	}

	req, err := tls_client_cffi.BuildRequest(requestInput.RequestInput)
	if err != nil {
		clientErr := tls_client_cffi.NewTLSClientError(err)
		return handleErrorResponse(sessionId, withSession, clientErr)
	}

	cookies := buildCookies(requestInput.RequestInput.RequestCookies)
	if len(cookies) > 0 {
		tlsClient.SetCookies(req.URL, cookies)
	}

	resp, reqErr := tlsClient.Do(req)
	if reqErr != nil {
		clientErr := tls_client_cffi.NewTLSClientError(fmt.Errorf("failed to do request: %w", reqErr))
		return handleErrorResponse(sessionId, withSession, clientErr)
	}

	if resp == nil {
		clientErr := tls_client_cffi.NewTLSClientError(fmt.Errorf("response is nil"))
		return handleErrorResponse(sessionId, withSession, clientErr)
	}

	targetCookies := tlsClient.GetCookies(resp.Request.URL)
	response, err := BuildResponse(sessionId, withSession, resp, targetCookies, requestInput.DetectEncoding)
	if err != nil {
		return handleErrorResponse(sessionId, withSession, err)
	}

	return &response
}

func handleResponse(logger *log.Logger) func(resp *http.Response, ctx *proxy.ProxyCtx) *http.Response {
	return func(resp *http.Response, ctx *proxy.ProxyCtx) *http.Response {
		logger.Printf("Modifying response from %s", resp.Request.URL)

		// Set your custom header
		resp.Header.Set("X-Returned-By", "My-Proxy")

		return resp
	}
}

func main() {
	flag.Parse()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	logger := log.New(os.Stdout, "", log.LstdFlags)

	proxyServer := proxy.NewProxyHttpServer()
	proxyServer.Verbose = true

	// Setup handlers
	proxyServer.OnRequest().HandleConnectFunc(func(host string, ctx *proxy.ProxyCtx) (*proxy.ConnectAction, string) {
		logger.Printf("Handling CONNECT request for host: %s", host)
		return proxy.AlwaysMitm(host, ctx)
	})
	proxyServer.OnRequest().DoFunc(handleRequest(logger))
	proxyServer.OnResponse().DoFunc(handleResponse(logger))

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{
		Addr:    addr,
		Handler: proxyServer,
	}

	// Start the server
	go func() {
		logger.Printf("Listening on http://0.0.0.0%s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Could not listen on %s: %v", addr, err)
		}
	}()

	// Wait for interrupt signal
	<-stop
	logger.Println("Shutting down the server...")

	// Shutdown the server gracefully
	if err := server.Shutdown(context.Background()); err != nil {
		logger.Fatalf("Server Shutdown Failed: %v", err)
	}

	logger.Println("Server gracefully stopped")
}
