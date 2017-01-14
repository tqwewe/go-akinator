package akinator

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// Used to check again a Response's Status.
const (
	StatusTimeout            = "KO - TIMEOUT"
	StatusIncorrectParameter = "KO - INCORRECT PARAMETER"
	StatusOk                 = "OK"
)

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

	resp, err := c.HTTPClient.Head("http://en.akinator.com/personnages/")
	if err != nil {
		return &c, err
	}
	resp.Body.Close()

	resp, err = c.HTTPClient.Get("http://api-us4.akinator.com/ws/new_session?" + url.Values{
		"partner":    {"1"},
		"player":     {"desktopPlayer"},
		"constraint": {"ETAT<>'AV'"},
	}.Encode())
	if err != nil {
		return &c, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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
