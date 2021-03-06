package main

import (
	"errors"
	"github.com/gorilla/context"
	"net/http"
	"regexp"
	"github.com/Sirupsen/logrus"
)

// GranularAccessMiddleware will check if a URL is specifically enabled for the key
type GranularAccessMiddleware struct {
	TykMiddleware
}

type GranularAccessMiddlewareConfig struct {}

func (m *GranularAccessMiddleware) New() {}

// GetConfig retrieves the configuration from the API config - we user mapstructure for this for simplicity
func (m *GranularAccessMiddleware) GetConfig() (interface{}, error) {
	return nil, nil
}

// ProcessRequest will run any checks on the request on the way through the system, return an error to have the chain fail
func (m *GranularAccessMiddleware) ProcessRequest(w http.ResponseWriter, r *http.Request, configuration interface{}) (error, int) {
	thisSessionState := context.Get(r, SessionData).(SessionState)
	authHeaderValue := context.Get(r, AuthHeaderValue).(string)
	
	sessionVersionData, foundAPI := thisSessionState.AccessRights[m.Spec.APIID]
	
	if foundAPI == false {
		log.Debug("Version not found")
		return nil, 200
	}
	
	if sessionVersionData.AllowedURLs == nil {
		log.Debug("No allowed URLS")
		return nil, 200
	}
	
	for url_regex, accessSpec := range(sessionVersionData.AllowedURLs) {
		log.Debug("Checking: ", r.URL.Path)
		log.Debug("Against: ", url_regex)
		asRegex, regexpErr := regexp.Compile(url_regex)
		
		if regexpErr != nil {
			log.Error("Regex error: ", regexpErr)
			return nil, 200
		}
		
		match := asRegex.MatchString(r.URL.Path)	
		if match {
			log.Debug("Match!")
			for _, method := range(accessSpec.Methods) {
				if method == r.Method {
					return nil, 200
				}
			}
		}
	}
	
	// No paths matched, disallow
	log.WithFields(logrus.Fields{
		"path":      r.URL.Path,
		"origin":    r.RemoteAddr,
		"key":       authHeaderValue,
		"api_found": false,
	}).Info("Attempted access to unauthorised endpoint (Granular).")

	return errors.New("Access to this resource has been disallowed"), 403

	
}
