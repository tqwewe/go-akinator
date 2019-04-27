package akinator

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

// Used to check again a Response's Status.
const (
	StatusTimeout            = "KO - TIMEOUT"
	StatusIncorrectParameter = "KO - INCORRECT PARAMETER"
	StatusOk                 = "OK"
)

// Domain info
const domainURL = "http://en.akinator.com"

// Client is used for each session with the Akinator.
type Client struct {
	HTTPClient     *http.Client
	apiURL         string
	responses      chan *Response
	identification struct {
		step      int
		session   string
		signature string
	}
	previousProgress float64
}

// NewClient returns a new *Client to use with the Akinator.
func NewClient() (*Client, error) {
	var c Client

	cookieJar, _ := cookiejar.New(nil)

	httpClient := &http.Client{
		Jar: cookieJar,
	}

	c.HTTPClient = httpClient
	c.responses = make(chan *Response, 1)

	// Get PHP session cookie and apiURL
	resp, err := c.HTTPClient.Get(domainURL)
	if err != nil {
		return &c, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &c, err
	}
	resp.Body.Close()

	c.apiURL, err = getAPIUrl(body)
	if err != nil {
		return &c, err
	}

	// Get uid_ext_session
	resp, err = c.HTTPClient.Get(domainURL + "/game")
	if err != nil {
		return &c, err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return &c, err
	}

	resp.Body.Close()

	uidExtSession, err := getUIDExtSession(body)
	if err != nil {
		return &c, err
	}

	frontAddr, err := getFrontAddr(body)
	if err != nil {
		return &c, err
	}

	// Begin session
	resp, err = c.HTTPClient.Get(c.apiURL + "/ws/new_session?" + url.Values{
		"partner":         {"1"},
		"player":          {"website-desktop"},
		"uid_ext_session": {uidExtSession},
		"frontaddr":       {frontAddr},
		"constraint":      {"ETAT<>'AV'"},
		"soft_constraint": {""},
		"question_filter": {""},
	}.Encode())
	if err != nil {
		return &c, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return &c, err
	}

	r, err := c.getResponse(body)
	if err != nil {
		return &c, err
	}

	c.responses <- r

	return &c, nil
}

// Next is used to recieve the next response in the channel.
func (c *Client) Next() <-chan *Response {
	return c.responses
}

var apuURLExp = regexp.MustCompile(`"urlWs":"(.*)",`)

func getAPIUrl(content []byte) (string, error) {
	matches := apuURLExp.FindSubmatch(content)
	if len(matches) < 2 || len(matches[1]) == 0 {
		return "", errors.New("failed to find api url")
	}

	return strings.Replace(string(matches[1]), "\\", "", -1), nil
}

var uidExtSessionExp = regexp.MustCompile(`uid_ext_session\s*=\s*\'(.*)\'`)

func getUIDExtSession(content []byte) (string, error) {
	matches := uidExtSessionExp.FindSubmatch(content)
	if len(matches) < 2 || len(matches[1]) == 0 {
		return "", errors.New("failed to find uid ext session")
	}

	return string(matches[1]), nil
}

var getFrontAddrExp = regexp.MustCompile(`frontaddr\s*=\s*\'(.*)\'`)

func getFrontAddr(content []byte) (string, error) {
	matches := getFrontAddrExp.FindSubmatch(content)
	if len(matches) < 2 || len(matches[1]) == 0 {
		return "", errors.New("failed to find front addr")
	}

	return string(matches[1]), nil
}
