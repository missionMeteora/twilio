# Twilio [![GoDoc](https://godoc.org/github.com/missionMeteora/twilio?status.svg)](https://godoc.org/github.com/missionMeteora/twilio) ![Status](https://img.shields.io/badge/status-beta-yellow.svg)

Twilio is a library which assists with sending SMS using the Twilio Api.

## Usage
``` go
package main

import (
	"github.com/missionMeteora/twilio"
)

func main() {
	tc := twilio.New([Twilio Account SID], [Twilio Auth Token], [From phone number])
	tc.Send("+15555555", "Hello, from Twilio!")
}
```