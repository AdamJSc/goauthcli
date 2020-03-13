package token

import (
	"context"
	"fmt"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
)

const (
	// defaultListenerPort is the default local port on which to listen for a request from the authorisation server
	defaultListenerPort = 3000

	// defaultListenerPath is the default path at which our auth code request handler will run
	defaultListenerPath = "/callback"
)

func GetDefaultCallbackURL() string {
	return fmt.Sprintf("http://localhost:%d%s", defaultListenerPort, defaultListenerPath)
}

// AuthCodeRequestHandler defines the signature for handling the auth code
// issued by the authorisation server and retrieving an access token
type AuthCodeRequestHandler func(r *http.Request) (oauth2.Token, error)

// OauthTokenExchanger configures the web server and request handler that run locally
// and listen for a response from the authorisation server
type OauthTokenExchanger struct {
	AuthServerRequestURL string
	Handler              AuthCodeRequestHandler
	ListenerPort         int
	ListenerPath         string
}

// ExchangeForToken launches our local web server and runs our request handler to exchange our auth code for an access token
func (o OauthTokenExchanger) TokenExchange() (oauth2.Token, error) {
	// launch in default browser
	browser.OpenURL(o.AuthServerRequestURL)

	// instantiate server and handler
	m := http.NewServeMux()
	s := http.Server{Addr: fmt.Sprintf(":%d", o.ListenerPort), Handler: m}

	// define token and error for reference within scope of callback func
	var t oauth2.Token
	var e error
	pT := &t
	pE := &e

	// full server callback handler
	cb := func(w http.ResponseWriter, r *http.Request) {
		// execute auth code handler
		token, err := o.Handler(r)

		// assign return values to parent scope via pointers
		*pT = token
		*pE = err

		// write to browser window
		if err != nil {
			http.Error(w, fmt.Sprintf("Error retrieving token: %s", err.Error()), http.StatusInternalServerError)
		} else {
			w.Write([]byte("Token retrieved successfully! Please close the browser window"))
		}

		// shutdown server
		go func() {
			// server has served its purpose, no need to check errors and shutdown gracefully
			s.Shutdown(context.Background())
		}()
	}

	// configure callback path
	path := strings.Trim(o.ListenerPath, "/")
	m.HandleFunc(fmt.Sprintf("/%s", path), cb)

	// launch server and listen for callback
	serverErr := s.ListenAndServe()

	// if our server generates an error and we don't already have one
	// from our callback handler, then return our server error instead
	if e == nil && serverErr != http.ErrServerClosed {
		e = serverErr
	}

	return t, e
}

// NewOauthTokenExchanger returns a new NewOauthTokenExchanger that uses default listener values
func NewOauthTokenExchanger(authServerRequestURL string, handler AuthCodeRequestHandler) OauthTokenExchanger {
	return OauthTokenExchanger{
		AuthServerRequestURL: authServerRequestURL,
		Handler:              handler,
		ListenerPort:         defaultListenerPort,
		ListenerPath:         defaultListenerPath,
	}
}
