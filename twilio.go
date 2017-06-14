package twilio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	// Location of Twilio Messages API endpoint
	base        = "https://%s:%s@api.twilio.com/2010-04-01/"
	messagesLoc = base + "Accounts/%s/Messages.json"
	findNumber  = base + "Accounts/%s/AvailablePhoneNumbers/US/Local.json?InRegion=TX&SmsEnabled=true"
	buyNumber   = base + "Accounts/%s/IncomingPhoneNumbers/Local.json"
	getMessages = base + "Accounts/%s/Messages.json"

	// Content type being used for requests
	contentType = "application/x-www-form-urlencoded"
	// Template of body for API requests
	bodyTmpl = "To=%s&From=%s&Body=%s"
)

func New(key, token, fromPhone, responseTemplate string) *Client {
	return &Client{
		key:              key,
		token:            token,
		fromPhone:        fromPhone,
		responseTemplate: responseTemplate,
	}
}

// Client values:
// 	- key: Twilio Account SID
// 	- token: Auth token paired with the provided Account SID
// 	- fromPhone: Number to be used as the 'from' location for sending SMS,
//		this number must be registered with the associated Twilio account
type Client struct {
	key              string
	token            string
	fromPhone        string
	responseTemplate string
}

type SMS struct {
	Sid         string `json:"sid"`
	DateCreated string `json:"date_created"`
	DateUpdate  string `json:"date_updated"`
	DateSent    string `json:"date_sent"`
	AccountSid  string `json:"account_sid"`
	To          string `json:"to"`
	From        string `json:"from"`
	Body        string `json:"body"`
	Direction   string `json:"direction"`
	Url         string `json:"uri"`
	Message     string `json:"message"`
}

func (sms *SMS) DateSentAsTime() time.Time {
	t, _ := time.Parse(time.RFC1123Z, sms.DateSent)
	return t
}

// Sends a basic (text-only) SMS with the 'to' value as the destination.
// Will return the error if there is an issue with:
//		- Sending the initial Api POST request
//		- The information provided to the Twilio Api
//		  (Invalid 'to' or 'from', invalid message, etc)
func (c *Client) Send(to, msg string) error {
	resp, err := http.Post(c.getUrl(messagesLoc), contentType, c.getBody(c.fromPhone, to, msg))
	if err != nil {
		return err
	}

	var r SMS
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return err
	}

	// If response object has a key of 'message', this means that
	// some sort of error has occured. Create a new error with the
	// provided message and return
	if r.Message != "" {
		return errors.New(r.Message)
	}

	return nil
}

// Sends a message with a specific from number
func (c *Client) SendWithNumber(from, to, msg string) error {
	resp, err := http.Post(c.getUrl(messagesLoc), contentType, c.getBody(from, to, msg))
	if err != nil {
		return err
	}

	var r SMS
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return err
	}

	// If response object has a key of 'message', this means that
	// some sort of error has occured. Create a new error with the
	// provided message and return
	if r.Message != "" {
		return errors.New(r.Message)
	}

	return nil
}

type Thread struct {
	Messages []*SMS `json:"messages"`
}

func (c *Client) GetThread(host, client string) ([]*SMS, error) {
	// Retrieves full SMS thread between two phone numbers

	// Retrieve all messages FROM the host to the client
	forwardThread, err := c.getThread(host, client)
	if err != nil {
		return nil, err
	}

	// Retreieve all messages FROM the client to the HOST
	reverseThread, err := c.getThread(client, host)
	if err != nil {
		return nil, err
	}

	messages := append(forwardThread.Messages, reverseThread.Messages...)

	// Lets sort thread by date
	sort.Slice(messages, func(i int, j int) bool {
		return messages[i].DateSentAsTime().Unix() > messages[j].DateSentAsTime().Unix()
	})

	return messages, nil
}

func (c *Client) getThread(from, to string) (*Thread, error) {
	formValues := url.Values{}
	if to != "" {
		formValues.Set("To", to)
	}

	if from != "" {
		formValues.Set("From", from)
	}

	// Retrieve all messages FROM the host to the client
	resp, err := http.Get(c.getUrl(getMessages) + "?" + formValues.Encode())
	if err != nil {
		return nil, err
	}

	var r Thread
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	// Retreieve all messages FROM the client to the HOST
	return &r, nil
}

type Numbers struct {
	Numbers []AvailableNumber `json:"available_phone_numbers"`
	Message string            `json:"message"`
}

type AvailableNumber struct {
	Sid                  string      `json:"sid"`
	AccountSid           string      `json:"account_sid"`
	FriendlyName         string      `json:"friendly_name"`
	PhoneNumber          string      `json:"phone_number"`
	VoiceURL             string      `json:"voice_url"`
	VoiceMethod          string      `json:"voice_method"`
	VoiceFallbackURL     interface{} `json:"voice_fallback_url"`
	VoiceFallbackMethod  string      `json:"voice_fallback_method"`
	StatusCallback       interface{} `json:"status_callback"`
	StatusCallbackMethod interface{} `json:"status_callback_method"`
	VoiceCallerIDLookup  interface{} `json:"voice_caller_id_lookup"`
	VoiceApplicationSid  interface{} `json:"voice_application_sid"`
	DateCreated          string      `json:"date_created"`
	DateUpdated          string      `json:"date_updated"`
	SmsURL               interface{} `json:"sms_url"`
	SmsMethod            string      `json:"sms_method"`
	SmsFallbackURL       interface{} `json:"sms_fallback_url"`
	SmsFallbackMethod    string      `json:"sms_fallback_method"`
	SmsApplicationSid    string      `json:"sms_application_sid"`
	Capabilities         struct {
		Voice bool `json:"voice"`
		Sms   bool `json:"sms"`
		Mms   bool `json:"mms"`
	} `json:"capabilities"`
	Beta       bool   `json:"beta"`
	APIVersion string `json:"api_version"`
	URI        string `json:"uri"`
	Message    string `json:"message"`
}

func (c *Client) AddNumber() (string, error) {
	// Creates a new phone number for given twilio account
	resp, err := http.Get(c.getUrl(findNumber))
	if err != nil {
		return "", err
	}

	var r Numbers
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return "", err
	}

	// If response object has a key of 'message', this means that
	// some sort of error has occured. Create a new error with the
	// provided message and return
	if r.Message != "" {
		return "", errors.New(r.Message)
	}

	var phoneNumber string
	if len(r.Numbers) != 0 && r.Numbers[0].PhoneNumber != "" {
		phoneNumber = r.Numbers[0].PhoneNumber
	}

	if phoneNumber == "" {
		return "", errors.New("No numbers available")
	}

	// Now buy the number!
	body := strings.NewReader(fmt.Sprintf("PhoneNumber=%s&SmsUrl=%s", phoneNumber, c.responseTemplate))
	resp, err = http.Post(c.getUrl(buyNumber), contentType, body)
	if err != nil {
		return "", err
	}

	var number AvailableNumber
	json.NewDecoder(resp.Body).Decode(&number)

	if number.Message != "" {
		return "", errors.New(number.Message)
	}

	return number.PhoneNumber, nil
}

func (c *Client) getUrl(url string) string {
	return fmt.Sprintf(url, c.key, c.token, c.key)
}

func (c *Client) getBody(from, to, msg string) io.Reader {
	return strings.NewReader(fmt.Sprintf(
		bodyTmpl,
		url.QueryEscape(to),
		url.QueryEscape(from),
		url.QueryEscape(msg),
	))
}
