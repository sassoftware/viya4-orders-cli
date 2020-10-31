// Copyright © 2020, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const viyaOrdersAPIHost string = "https://api.sas.com"
const viyaOrdersAPIBasePath string = "/mysas"
const viyaOrdersAPITokenPath string = "/token"

type apigeeToken struct {
	TokenType             string   `json:"token_type"`
	AccessToken           string   `json:"access_token"`
	IssuedAt              int      `json:"issued_at"`
	ExpiresIn             int      `json:"expires_in"`
	Scope                 string   `json:"scope"`
}

// GetBearerToken calls the /token API endpoint to exchange client credentials for a Bearer token.
func GetBearerToken() (token string, err error) {
	data := url.Values{}
	id, err := base64.StdEncoding.DecodeString(viper.GetString("clientCredentialsId"))
	if err != nil {
		return token, errors.New("ERROR: attempt to decode clientCredentialsId failed: " + err.Error())
	}
	sec, err := base64.StdEncoding.DecodeString(viper.GetString("clientCredentialsSecret"))
	if err != nil {
		return token, errors.New("ERROR: attempt to decode clientCredentialsSecret failed: " + err.Error())
	}
	data.Set("client_id", string(id))
	data.Set("client_secret", string(sec))
	data.Set("grant_type", "client_credentials")

	// Build the request URL.
	u, err := url.ParseRequestURI(viyaOrdersAPIHost)
	if err != nil {
		return token, errors.New("ERROR: attempt to parse Bearer token request URI failed: " + err.Error())
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s", viyaOrdersAPIBasePath)
	fmt.Fprintf(&b, "%s", viyaOrdersAPITokenPath)
	u.Path = b.String()
	urlStr := u.String()

	client := &http.Client{}
	r, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return token, errors.New("ERROR: setup of Bearer token request failed: " + err.Error())
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		return token, errors.New("ERROR: Bearer token request failed to complete: " + err.Error())
	}

	// Get the response.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return token, errors.New("ERROR: read of response body from Bearer token request failed: " + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		var em = "ERROR: Bearer token request failed: "
		var emErr string
		if len(body) > 0 {
			emErr = string(body)
		} else {
			emErr = fmt.Sprintf("%d -- %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return token, errors.New(em+emErr)
	}

	var apigeeToken apigeeToken
	err = json.Unmarshal(body, &apigeeToken)
	if err != nil {
		return token, errors.New("ERROR: unmarshalling of token API response failed: " + err.Error())
	}

	return apigeeToken.AccessToken, nil
}