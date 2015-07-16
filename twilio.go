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
	// Location of Twilio Messages API endpoint
	twilioLoc = "https://%s:%s@api.twilio.com/2010-04-01/Accounts/%s/Messages.json"
	// Content type being used for requests
	contentType = "application/x-www-form-urlencoded"
	// Template of body for API requests
	bodyTmpl = "To=%s&From=%s&Body=%s"
)

func New(key, token, fromPhone string) *Client {
	return &Client{
		key:       key,
		token:     token,
		fromPhone: fromPhone,
	}
}

// Client values:
// 	- key: Twilio Account SID
// 	- token: Auth token paired with the provided Account SID
// 	- fromPhone: Number to be used as the 'from' location for sending SMS,
//		this number must be registered with the associated Twilio account
type Client struct {
	key       string
	token     string
	fromPhone string
}

// Sends a basic (text-only) SMS with the 'to' value as the destination.
// Will return an error if there is an issue with:
//		- Sending the initial Api POST request
//		- The information provided to the Twilio Api
//		  (Invalid 'to' or 'from', invalid message, etc)
func (c *Client) Send(to, msg string) error {
	resp, err := http.Post(
		c.getUrl(),
		contentType,
		strings.NewReader(c.getBody(to, msg)),
	)
	if err != nil {
		return err
	}

	var r map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&r)

	// If response object has a key of 'message', this means that
	// some sort of error has occured. Create a new error with the
	// provided message and return
	if msg, ok := r["message"].(string); ok {
		return errors.New(msg)
	}

	return nil
}

func (c *Client) getUrl() string {
	return fmt.Sprintf(twilioLoc, c.key, c.token, c.key)
}

func (c *Client) getBody(to, msg string) string {
	return fmt.Sprintf(
		bodyTmpl,
		url.QueryEscape(to),
		url.QueryEscape(c.fromPhone),
		url.QueryEscape(msg),
	)
}
