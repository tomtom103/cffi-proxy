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

	http "github.com/bogdanfinn/fhttp"
	"github.com/tomtom103/cffi-proxy/goproxy"
	proxy "github.com/tomtom103/cffi-proxy/goproxy"
	mitm "github.com/tomtom103/cffi-proxy/goproxy/mitm"
)

var port int

func init() {
	flag.IntVar(&port, "port", 8000, "Port where the proxy will listen on")
}

type RequestHandler interface {
	handleConnect(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string)
	handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response)
	handleResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response
}

// Implementation of the RequestHandler interface
type MitmRequestHandler struct {
	transport mitm.RequestTransport
}

func (h *MitmRequestHandler) handleConnect(host string, ctx *proxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	return proxy.AlwaysMitm(host, ctx)
}

func (h *MitmRequestHandler) handleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	ctx.Logf("Handling request: %s %s", req.Method, req.URL.String())

	response, err := h.transport.HandleRequest(req, ctx)
	if err != nil {
		// Log and return an error response
		ctx.Warnf("Error handling request: %v", err)
		return req, &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(strings.NewReader(fmt.Sprintf("Proxy error: %v", err))),
			Request:    req,
		}
	}

	// If the transport successfully handles the request, return the response
	return nil, response
}

func (h *MitmRequestHandler) handleResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	ctx.Logf("Handling response: %s %d", resp.Request.URL.String(), resp.StatusCode)

	// Example: Add a custom header
	resp.Header.Set("X-Proxy-Handler", "MitmRequestHandler")

	// Example: Log response details
	ctx.Logf("Response Headers: %v", resp.Header)

	// Example: Modify the response body (if needed)
	if resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			ctx.Warnf("Error reading response body: %v", err)
		} else {
			ctx.Logf("Response body: %s", string(body))
		}
		resp.Body = io.NopCloser(strings.NewReader(string(body)))
	}

	return resp
}

func main() {
	flag.Parse()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	logger := log.New(os.Stdout, "", log.LstdFlags)

	proxyServer := proxy.NewProxyHttpServer()
	proxyServer.Verbose = true

	transport := &mitm.CffiRequestTransport{}
	handler := &MitmRequestHandler{transport: transport}

	// Setup handlers
	proxyServer.OnRequest().HandleConnectFunc(func(host string, ctx *proxy.ProxyCtx) (*proxy.ConnectAction, string) {
		ctx.Logf("Handling CONNECT request for host: %s", host)
		return handler.handleConnect(host, ctx)
	})
	proxyServer.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		// ctx.Logf("Handling CONNECT request for )
		return handler.handleRequest(req, ctx)
	})
	proxyServer.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		return handler.handleResponse(resp, ctx)
	})

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

	if err := server.Shutdown(context.Background()); err != nil {
		logger.Fatalf("Server Shutdown Failed: %v", err)
	}

	logger.Println("Server gracefully stopped")
}
