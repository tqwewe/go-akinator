package akinator

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
)

// Used to check again a Response's Status.
const (
	StatusTimeout            = "KO - TIMEOUT"
	StatusIncorrectParameter = "KO - INCORRECT PARAMETER"
	StatusOk                 = "OK"
)

// Domain info
const domainURL = "http://en.akinator.com"
const apiURL = "https://srv13.akinator.com:9196"

// Client is used for each session with the Akinator.
type Client struct {
	HTTPClient     *http.Client
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

	// Get PHP session cookie
	resp, err := c.HTTPClient.Head(domainURL)
	if err != nil {
		return &c, err
	}
	resp.Body.Close()

	// Get uid_ext_session
	resp, err = c.HTTPClient.Get(domainURL + "/game")
	if err != nil {
		return &c, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &c, err
	}

	resp.Body.Close()

	uidExtSession, err := c.getUidExtSession(body)
	if err != nil {
		return &c, err
	}

	frontAddr, err := c.getFrontAddr(body)
	if err != nil {
		return &c, err
	}

	// Begin session
	resp, err = c.HTTPClient.Get(apiURL + "/ws/new_session?" + url.Values{
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

var uidExtSessionExp = regexp.MustCompile(`uid_ext_session\s*=\s*\'(.*)\'`)

func (c *Client) getUidExtSession(content []byte) (string, error) {
	matches := uidExtSessionExp.FindSubmatch(content)
	if len(matches) < 2 || len(matches[1]) == 0 {
		return "", errors.New("failed to find uid ext session")
	}

	return string(matches[1]), nil
}

var getFrontAddrExp = regexp.MustCompile(`frontaddr\s*=\s*\'(.*)\'`)

func (c *Client) getFrontAddr(content []byte) (string, error) {
	matches := getFrontAddrExp.FindSubmatch(content)
	if len(matches) < 2 || len(matches[1]) == 0 {
		return "", errors.New("failed to find front addr")
	}

	return string(matches[1]), nil
}
