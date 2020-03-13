# goauthcli

This is a Go package that facilitates the OAuth2 workflow for a command-line program.

It abstracts the process of running a local web server, listening for a callback/auth code
from the authorisation server, and exchanging the auth code for a usable access token.

Existing examples of this process with respect to the command-line are generally based around
using the resultant access token within the callback handler itself.

This not only enforces subsequent business logic to be undesirably nested, but also requires that
the local web server continues running for the duration of the command-line program.

The `goauthcli` package encapsulates this process such that the access token is returned to the
parent scope following the exchange, so the web server can be shutdown.

## Usage

This package was written to facilitate the integration of the Spotify Web API within the context
of a command-line program.

The following example has therefore been written around this use-case.

### Config

The main component is the `token.OauthTokenExchanger` type, which comprises the following data fields:

* `AuthServerRequestURL` - string representing the authorisation server's URL for generating an auth code
* `Handler`              - function to act as our token callback handler which 
                           accepts a HTTP request sent from the authorisation server
                           (auth code is part of the query string), and returns an access token
                           along with an error
* `ListenerPort`         - port number for the local web server to listen on
* `ListenerPath`         - path/route for our local web server callback handler

You can instantiate a new exchanger using the method `token.NewOauthTokenExchanger()` which will default
`ListenerPort` to `3000` and `ListenerPath` to `/callback`.

### Prerequisites

1. A new client app has been created at `https://developer.spotify.com/`
2. There are no other applications listening on port `3000`
(this is the package's default but can be overridden as alluded to above)

### Example

```go
package main

import (
	"github.com/adamjsc/goauthcli/token"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

func main() {
	// get these values from your developer dashboard
	spotifyClientID := "XXX"
	spotifyClientSecret := "XXX"

	// see https://github.com/zmb3/spotify for docs
	auth := spotify.NewAuthenticator(token.GetDefaultCallbackURL(), spotify.ScopeUserReadPrivate)
	auth.SetAuthInfo(spotifyClientID, spotifyClientSecret)

	state := "spotify_example_state" // this value is an additional validation check and can be literally anything 
	url := auth.AuthURL(state)

	// this is our function that will handle the callback from the Spotify authorisation server
	requestHandler := func(r *http.Request) (oauth2.Token, error) {
		token, err := auth.Token(state, r)
		if err != nil {
			return oauth2.Token{}, err
		}

		return *token, err
	}

	// get our access token
	exchanger := token.NewOauthTokenExchanger(url, requestHandler)
	token, err := exchanger.TokenExchange()
	if err != nil {
		log.Fatal(err)
	}

	// see https://github.com/zmb3/spotify for docs
	client := auth.NewClient(&token)
	user, err := client.CurrentUser()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Hi %s!\nYour Spotify URI is: %s\n", user.DisplayName, user.URI)
}
```
