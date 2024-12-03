package mitm

import (
	"fmt"
	"io"
	"strings"

	http "github.com/bogdanfinn/fhttp"
	tls_client_cffi "github.com/bogdanfinn/tls-client/cffi_src"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/tomtom103/cffi-proxy/goproxy"
	"golang.org/x/exp/rand"
)

type RequestTransport interface {
	HandleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error)
}

type ExtendedRequestInput struct {
	tls_client_cffi.RequestInput
	WantHistory    bool `json:"wantHistory"`
	DetectEncoding bool `json:"detectEncoding"`
}

// Implement the RequestTransport interface
type CffiRequestTransport struct{}

func (t *CffiRequestTransport) HandleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
	ctx.Logf("Handling request for %s", req.URL)

	// Create the extended request input accepted by the CFFI client
	requestInput := tls_client_cffi.RequestInput{
		RequestMethod:       req.Method,
		RequestUrl:          req.URL.String(),
		Headers:             make(map[string]string),
		WithDebug:           true,
		FollowRedirects:     true,
		TLSClientIdentifier: t.getRandomProfileString(req, ctx),
	}

	// Copy headers from the incoming request
	for name, values := range req.Header {
		requestInput.Headers[name] = values[0]
	}

	tlsClient, sessionId, withSession, err := tls_client_cffi.CreateClient(requestInput)
	if err != nil {
		return &http.Response{}, err
	}

	cffi_req, err := tls_client_cffi.BuildRequest(requestInput)
	if err != nil {
		clientErr := tls_client_cffi.NewTLSClientError(err)
		return &http.Response{}, clientErr
	}

	cookies := t.buildCookies(requestInput.RequestCookies)
	if len(cookies) > 0 {
		tlsClient.SetCookies(cffi_req.URL, cookies)
	}

	resp, reqErr := tlsClient.Do(cffi_req)
	if reqErr != nil {
		clientErr := tls_client_cffi.NewTLSClientError(fmt.Errorf("failed to do request: %w", reqErr))
		return &http.Response{}, clientErr
	}

	if resp == nil {
		clientErr := tls_client_cffi.NewTLSClientError(fmt.Errorf("response is nil"))
		return &http.Response{}, clientErr
	}

	targetCookies := tlsClient.GetCookies(resp.Request.URL)

	cffi_resp, err := tls_client_cffi.BuildResponse(sessionId, withSession, resp, targetCookies, requestInput)
	if err != nil {
		return &http.Response{}, err
	}

	response := &http.Response{
		StatusCode: cffi_resp.Status,
		Header:     cffi_req.Header,
		Body:       io.NopCloser(strings.NewReader(cffi_resp.Body)),
		Request:    req,
	}

	return response, nil
}

func (t *CffiRequestTransport) getRandomProfileString(req *http.Request, ctx *goproxy.ProxyCtx) string {
	// TODO: Pick based on incoming request
	keys := make([]string, 0, len(profiles.MappedTLSClients))
	for k := range profiles.MappedTLSClients {
		keys = append(keys, k)
	}
	randIndex := rand.Intn(len(keys))
	randomKey := keys[randIndex]
	return randomKey
}

func (t *CffiRequestTransport) buildCookies(cookies []tls_client_cffi.Cookie) []*http.Cookie {
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
