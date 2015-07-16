package twilio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	twilioLoc   = "https://%s:%s@api.twilio.com/2010-04-01/Accounts/%s/Messages.json"
	contentType = "application/x-www-form-urlencoded"
	bodyTmpl    = "To=%s&From=%s&Body=%s"
)

func New(key, token string) *Client {
	return &Client{
		key:   key,
		token: token,
	}
}

type Client struct {
	key   string
	token string
}

func (c *Client) Send(to, from, msg string) error {
	resp, err := http.Post(
		c.getUrl(),
		contentType,
		strings.NewReader(c.getBody(to, from, msg)),
	)
	if err != nil {
		return err
	}

	var r map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&r)

	if msg, ok := r["message"].(string); ok {
		return errors.New(msg)
	}

	return nil
}

func (c *Client) getUrl() string {
	return fmt.Sprintf(twilioLoc, c.key, c.token, c.key)
}

func (c *Client) getBody(to, from, msg string) string {
	return fmt.Sprintf(
		bodyTmpl,
		url.QueryEscape(to),
		url.QueryEscape(from),
		url.QueryEscape(msg),
	)
}
