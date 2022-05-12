package hubspot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)

// ClientConfig object used for client creation
type ClientConfig struct {
	APIHost     string
	APIKey      string
	OAuthToken  string
	HTTPTimeout time.Duration
	DialTimeout time.Duration
	TLSTimeout  time.Duration
}

// NewClientConfig constructs a ClientConfig object with the environment variables set as default
func NewClientConfig() ClientConfig {
	apiHost := "https://api.hubapi.com"
	var apiKey string
	var oauthToken string

	if os.Getenv("HUBSPOT_API_HOST") != "" {
		apiHost = os.Getenv("HUBSPOT_API_HOST")
	}
	if os.Getenv("HUBSPOT_API_KEY") != "" {
		apiKey = os.Getenv("HUBSPOT_API_KEY")
	}
	if os.Getenv("HUBSPOT_OAUTH_TOKEN") != "" {
		oauthToken = os.Getenv("HUBSPOT_OAUTH_TOKEN")
	}

	return ClientConfig{
		APIHost:     apiHost,
		APIKey:      apiKey,
		OAuthToken:  oauthToken,
		HTTPTimeout: 10 * time.Second,
		DialTimeout: 5 * time.Second,
		TLSTimeout:  5 * time.Second,
	}
}

// Client object
type Client struct {
	config ClientConfig
}

// NewClient constructor
func NewClient(config ClientConfig) Client {
	return Client{
		config: config,
	}
}

// addAPIKey adds HUBSPOT_API_KEY param to a given URL.
func (c Client) addAPIKey(uri url) (url, error) {
	if c.config.APIKey != "" {
/*		uri, err := url.Parse(u)
		if err != nil {
			return u, err
		} */
		q := uri.Query()
		q.Set("hapikey", c.config.APIKey)
		uri.RawQuery = q.Encode()
//		u = uri.String()
	}

	return uri, nil
}

// Request executes any HubSpot API method using the current client configuration
func (c Client) Request(method, endpoint string, data, response interface{}) error {
	// Construct endpoint URL
	u, err := url.Parse(c.config.APIHost)
	uri3 := u.String()
	if err != nil {
		return fmt.Errorf("hubspot.Client.Request(): url.Parse(): %v", err)
	}
	pattern := regexp.MustCompile("([^?]+)(\?(.*))?")
	matches := regexp.FindStringSubmatch(endpoint, -1)
	ep_path := matches[0][1]
	ep_variables := matches[0][3]
	u.Path = path.Join(u.Path, ep_path)

	q := u.Query()
	for pair := range regexp.MustCompile("&").Split(ep_variables, -1) {
//	for pair := ep_variables.Split() {
		pattern := regexp.MustCompile("(\w+)=(\w+)")
		matches := regexp.FindStringSubmatch(pair, -1)
		q.Set(matches[0][1], matches[0][2])
	}
	u.RawQuery = q.Encode()

	uri4 := u.String()

	// API Key authentication
	if c.config.APIKey != "" {
//		uri, err = c.addAPIKey(uri)
		u, err = c.addAPIKey(u)
		if err != nil {
			return fmt.Errorf("hubspot.Client.Request(): c.addAPIKey(): %v", err)
		}
	}
	uri := u.String()
	uri2 := uri

	// Init request object
	var req *http.Request

	// Send data?
	if data != nil {
		// Encode data to JSON
		dataEncoded, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("hubspot.Client.Request(): json.Marshal(): %v", err)
		}
		buf := bytes.NewBuffer(dataEncoded)

		// Create request
		req, err = http.NewRequest(method, uri, buf)
	} else {
		// Create no-data request
		req, err = http.NewRequest(method, uri, nil)
	}
	if err != nil {
		return fmt.Errorf("hubspot.Client.Request(): http.NewRequest(): %v", err)
	}

	// OAuth authentication
	if c.config.APIKey == "" && c.config.OAuthToken != "" {
		req.Header.Add("Authorization", "Bearer "+c.config.OAuthToken)
	}

	// Headers
	req.Header.Add("Content-Type", "application/json")

	// Execute and read response body
	netClient := &http.Client{
		Timeout: c.config.HTTPTimeout,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: c.config.DialTimeout,
			}).Dial,
			TLSHandshakeTimeout: c.config.TLSTimeout,
		},
	}
	resp, err := netClient.Do(req)
	if err != nil {
		return fmt.Errorf("hubspot.Client.Request(): c.config.HTTPClient.Do(): %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("hubspot.Client.Request(): ioutil.ReadAll(): %v", err)
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("hubspot.Client.Request(): HTTP Error fetching '%s' (%s || %s): %v \n%s\nuri2: %s\nuri3: %s\nuri4: %s\nendpoint: %s\n", uri, u.String(), c.config.APIHost, err, string(body), uri2, uri3, uri4, endpoint)
	}

	// Get data?
	if response != nil {
		err = json.Unmarshal(body, &response)
		if err != nil {
			return fmt.Errorf("hubspot.Client.Request(): json.Unmarshal(): %v \n%s", err, string(body))
		}
	}

	// Return HTTP errors
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return fmt.Errorf("HubSpot API error: %d - %s \n%s", resp.StatusCode, resp.Status, string(body))
	}

	// Done!
	return nil
}

// Generate a path for an object, optionally with a version
func (c Client) objectPath(object string, path string, version string) (op string) {
	if version == "" {
		version = "v3"
	}
	rv := ""
	switch version {
		case "v3":
			rv = fmt.Sprintf("/crm/%s/objects/%s", version, object)
			if path != "" {
				rv += "/" + path
			}
		default:
			fmt.Errorf("Version not implemented yet: " + version)
	}

	return rv
}
