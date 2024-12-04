package uagen

// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io/ioutil"
// 	"path/filepath"
// 	"strconv"
// 	"strings"
// )

// var pascalizeUpper = map[string]struct{}{
// 	"dnt": {},
// 	"rtt": {},
// 	"ect": {},
// }

// func pascalize(name string) string {
// 	if strings.HasPrefix(name, ":") || strings.HasPrefix(name, "sec-ch-ua") {
// 		return name
// 	}
// 	if _, ok := pascalizeUpper[name]; ok {
// 		return strings.ToUpper(name)
// 	}
// 	return strings.Title(strings.ToLower(name))
// }

// func pascalizeHeaders(headers map[string]string) map[string]string {
// 	pascalHeaders := make(map[string]string)
// 	for k, v := range headers {
// 		pascalHeaders[pascalize(k)] = v
// 	}
// 	return pascalHeaders
// }

// func getUserAgent(headers map[string]string) (string, error) {
// 	if ua, ok := headers["User-Agent"]; ok {
// 		return ua, nil
// 	}
// 	if ua, ok := headers["user-agent"]; ok {
// 		return ua, nil
// 	}
// 	return "", errors.New("User-Agent header not found")
// }

// func getBrowser(userAgent string) (string, error) {
// 	switch {
// 	case strings.Contains(userAgent, "Firefox"):
// 		return "firefox", nil
// 	case strings.Contains(userAgent, "Chrome"):
// 		return "chrome", nil
// 	case strings.Contains(userAgent, "Safari"):
// 		return "safari", nil
// 	case strings.Contains(userAgent, "Edge"):
// 		return "edge", nil
// 	default:
// 		return "", errors.New("Browser not recognized in User-Agent")
// 	}
// }

// func tuplify(obj interface{}) []string {
// 	if obj == nil {
// 		return nil
// 	}
// 	switch v := obj.(type) {
// 	case string:
// 		return []string{v}
// 	case []string:
// 		return v
// 	case []interface{}:
// 		var result []string
// 		for _, item := range v {
// 			if s, ok := item.(string); ok {
// 				result = append(result, s)
// 			}
// 		}
// 		return result
// 	default:
// 		return nil
// 	}
// }

// type Browser struct {
// 	Name        string
// 	MinVersion  *int
// 	MaxVersion  *int
// 	HttpVersion string
// }

// type HttpBrowserObject struct {
// 	Name           string
// 	Version        []int
// 	CompleteString string
// 	HttpVersion    string
// }

// func (hbo *HttpBrowserObject) IsHttp2() bool {
// 	return hbo.HttpVersion == "2"
// }

// type HeaderGenerator struct {
// 	RelaxationOrder        []string
// 	InputGeneratorNetwork  *BayesianNetwork
// 	HeaderGeneratorNetwork *BayesianNetwork
// 	Options                map[string]interface{}
// 	UniqueBrowsers         []HttpBrowserObject
// 	HeadersOrder           map[string][]string
// }

// func NewHeaderGenerator(dataDir string, options map[string]interface{}) (*HeaderGenerator, error) {
// 	hg := &HeaderGenerator{
// 		RelaxationOrder: []string{"locales", "devices", "operatingSystems", "browsers"},
// 		Options:         options,
// 	}

// 	// Load the Bayesian networks
// 	inputNetworkPath := filepath.Join(dataDir, "input-network.zip")
// 	headerNetworkPath := filepath.Join(dataDir, "header-network.zip")

// 	var err error
// 	hg.InputGeneratorNetwork, err = NewBayesianNetwork(inputNetworkPath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	hg.HeaderGeneratorNetwork, err = NewBayesianNetwork(headerNetworkPath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Load unique browsers and headers order
// 	hg.UniqueBrowsers, err = hg.loadUniqueBrowsers(filepath.Join(dataDir, "browser-helper-file.json"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	hg.HeadersOrder, err = hg.loadHeadersOrder(filepath.Join(dataDir, "headers-order.json"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	return hg, nil
// }

// func (hg *HeaderGenerator) Generate(overrides map[string]interface{}) (map[string]string, error) {
// 	options := hg.mergeOptions(overrides)

// 	headers, err := hg.getHeaders(options)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Convert headers to the correct case based on HTTP version
// 	if options["http_version"] == "2" {
// 		return pascalizeHeaders(headers), nil
// 	}
// 	return headers, nil
// }

// func (hg *HeaderGenerator) getHeaders(options map[string]interface{}) (map[string]string, error) {
// 	possibleAttributeValues := hg.getPossibleAttributeValues(options)

// 	// Generate constraints for the Bayesian network
// 	constraints := hg.prepareConstraints(possibleAttributeValues)

// 	// Generate input sample
// 	inputSample, err := hg.InputGeneratorNetwork.GenerateConsistentSampleWhenPossible(constraints)
// 	if err != nil || inputSample == nil {
// 		return nil, errors.New("failed to generate headers based on the input constraints")
// 	}

// 	// Generate headers sample
// 	generatedSample, err := hg.HeaderGeneratorNetwork.GenerateSample(inputSample)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Process the generated sample to produce headers
// 	headers := hg.processGeneratedSample(generatedSample, options)

// 	return hg.orderHeaders(headers), nil
// }

// func (hg *HeaderGenerator) mergeOptions(overrides map[string]interface{}) map[string]interface{} {
// 	merged := make(map[string]interface{})
// 	for k, v := range hg.Options {
// 		merged[k] = v
// 	}
// 	for k, v := range overrides {
// 		merged[k] = v
// 	}
// 	return merged
// }

// func (hg *HeaderGenerator) loadUniqueBrowsers(filePath string) ([]HttpBrowserObject, error) {
// 	data, err := ioutil.ReadFile(filePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var uniqueBrowserStrings []string
// 	if err := json.Unmarshal(data, &uniqueBrowserStrings); err != nil {
// 		return nil, err
// 	}

// 	var uniqueBrowsers []HttpBrowserObject
// 	for _, browserStr := range uniqueBrowserStrings {
// 		if browserStr != "*MISSING_VALUE*" {
// 			hbo, err := hg.prepareHttpBrowserObject(browserStr)
// 			if err != nil {
// 				return nil, err
// 			}
// 			uniqueBrowsers = append(uniqueBrowsers, hbo)
// 		}
// 	}
// 	return uniqueBrowsers, nil
// }

// func (hg *HeaderGenerator) loadHeadersOrder(filePath string) (map[string][]string, error) {
// 	data, err := ioutil.ReadFile(filePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var headersOrder map[string][]string
// 	if err := json.Unmarshal(data, &headersOrder); err != nil {
// 		return nil, err
// 	}

// 	return headersOrder, nil
// }

// func (hg *HeaderGenerator) processGeneratedSample(generatedSample map[string]interface{}, options map[string]interface{}) map[string]string {
// 	headers := make(map[string]string)

// 	// Convert generatedSample to headers map
// 	for k, v := range generatedSample {
// 		if s, ok := v.(string); ok {
// 			headers[k] = s
// 		}
// 	}

// 	// Add Accept-Language header
// 	locales, _ := options["locales"].([]string)
// 	acceptLanguage := hg.getAcceptLanguageHeader(locales)
// 	if acceptLanguage != "" {
// 		headers["Accept-Language"] = acceptLanguage
// 	}

// 	// Add Sec-Fetch headers if necessary
// 	// Implement logic based on the browser and version

// 	// Remove unwanted headers
// 	for k, v := range headers {
// 		if strings.HasPrefix(k, "*") || v == "*MISSING_VALUE*" {
// 			delete(headers, k)
// 		}
// 	}

// 	return headers
// }

// func (hg *HeaderGenerator) getAcceptLanguageHeader(locales []string) string {
// 	var parts []string
// 	for i, locale := range locales {
// 		q := 1.0 - float64(i)*0.1
// 		if q <= 0 {
// 			q = 0.1
// 		}
// 		parts = append(parts, fmt.Sprintf("%s;q=%.1f", locale, q))
// 	}
// 	return strings.Join(parts, ", ")
// }

// func (hg *HeaderGenerator) prepareHttpBrowserObject(browserStr string) (HttpBrowserObject, error) {
// 	parts := strings.Split(browserStr, "|")
// 	if len(parts) != 2 {
// 		return HttpBrowserObject{}, errors.New("invalid browser string format")
// 	}
// 	browserInfo, httpVersion := parts[0], parts[1]
// 	if browserInfo == "*MISSING_VALUE*" {
// 		return HttpBrowserObject{CompleteString: "*MISSING_VALUE*", HttpVersion: httpVersion}, nil
// 	}
// 	subParts := strings.Split(browserInfo, "/")
// 	if len(subParts) != 2 {
// 		return HttpBrowserObject{}, errors.New("invalid browser info format")
// 	}
// 	name, versionStr := subParts[0], subParts[1]
// 	versionParts := strings.Split(versionStr, ".")
// 	var version []int
// 	for _, vp := range versionParts {
// 		v, err := strconv.Atoi(vp)
// 		if err != nil {
// 			return HttpBrowserObject{}, err
// 		}
// 		version = append(version, v)
// 	}
// 	return HttpBrowserObject{Name: name, Version: version, CompleteString: browserStr, HttpVersion: httpVersion}, nil
// }
