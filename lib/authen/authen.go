//Copyright © 2020, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
//SPDX-License-Identifier: Apache-2.0

package authen

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
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

//Call the /token endpoint to exchange client credentials for Bearer token.
func GetBearerToken() string {
	data := url.Values{}
	id, err := base64.StdEncoding.DecodeString(viper.GetString("clientCredentialsId"))
	if err != nil {
		log.Panic("Error decoding clientCredentialsId: " + err.Error())
	}
	sec, err := base64.StdEncoding.DecodeString(viper.GetString("clientCredentialsSecret"))
	if err != nil {
		log.Panic("Error decoding clientCredentialsSecret: " + err.Error())
	}
	data.Set("client_id", string(id))
	data.Set("client_secret", string(sec))
	data.Set("grant_type", "client_credentials")

	//Build the request URL.
	u, err := url.ParseRequestURI(viyaOrdersAPIHost)
	if err != nil {
		log.Panic("Error parsing Bearer token request URI: " + err.Error())
	}
	//u.Path = viyaOrdersAPIBasePath + viyaOrdersAPITokenPath
	var b strings.Builder
	fmt.Fprintf(&b, "%s", viyaOrdersAPIBasePath)
	fmt.Fprintf(&b, "%s", viyaOrdersAPITokenPath)
	u.Path = b.String()
	urlStr := u.String()

	client := &http.Client{}
	r, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		log.Panic("Error setting up Bearer token request: " + err.Error())
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(r)
	if err != nil {
		log.Panic("Error on Bearer token request: " + err.Error())
	}

	//Get the response.
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic("Error reading response body from Bearer token request: " + err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		var em = fmt.Sprintf("Bearer token request returned: %d -- %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		if body != nil && len(body) > 0 {
			em += fmt.Sprintf(". Error: %s", string(body))
		}
		log.Panic(em)
	}

	var apigeeToken apigeeToken
	err = json.Unmarshal(body, &apigeeToken)
	if err != nil {
		log.Panic("Error unmarshalling token API token response: " + err.Error())
	}

	return apigeeToken.AccessToken
}