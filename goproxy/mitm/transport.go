package mitm

import (
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	profiles "github.com/bogdanfinn/tls-client/profiles"
	"github.com/tomtom103/cffi-proxy/goproxy"
	"golang.org/x/exp/rand"
)

type RequestTransport interface {
	HandleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error)
}

// Implement the RequestTransport interface
type CffiRequestTransport struct{}

func (t *CffiRequestTransport) getRandomProfileString(req *http.Request, ctx *goproxy.ProxyCtx) profiles.ClientProfile {
	// TODO: Pick based on incoming request
	keys := make([]string, 0, len(profiles.MappedTLSClients))
	for k := range profiles.MappedTLSClients {
		keys = append(keys, k)
	}
	randIndex := rand.Intn(len(keys))
	randomKey := keys[randIndex]
	return profiles.MappedTLSClients[randomKey]
}

func (t *CffiRequestTransport) createHttpClient(req *http.Request, ctx *goproxy.ProxyCtx) (tls_client.HttpClient, error) {
	timeoutOptions := tls_client.WithTimeoutSeconds(tls_client.DefaultTimeoutSeconds)
	transportOptions := &tls_client.TransportOptions{
		DisableCompression: true,
		DisableKeepAlives:  true,
	}
	options := []tls_client.HttpClientOption{
		timeoutOptions,
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithClientProfile(t.getRandomProfileString(req, ctx)),
		tls_client.WithCatchPanics(),
		tls_client.WithTransportOptions(transportOptions),
	}

	/*
		If proxy configured:

		if proxy != nil && *proxy != "" {
			options = append(options, tls_client.WithProxyUrl(*proxy))
		}
	*/
	tlsClient, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)

	return tlsClient, err
}

func (t *CffiRequestTransport) HandleRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Response, error) {
	ctx.Logf("Handling request for %s", req.URL)

	// Apparently we can't have this set for client requests for some reason
	req.RequestURI = ""

	tlsClient, err := t.createHttpClient(req, ctx)
	if err != nil {
		return &http.Response{}, err
	}

	response, err := tlsClient.Do(req)

	if err != nil || response == nil {
		return &http.Response{}, err
	}

	return response, nil
}
