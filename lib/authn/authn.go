// Copyright © 2020-2023, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package authn provides a func that will exchange OAuth client credentials for a Bearer token that will expire after
// 30 minutes.
package authn

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	viyaOrdersAPIHost      string = "https://api.sas.com"
	viyaOrdersAPIBasePath  string = "/mysas"
	viyaOrdersAPITokenPath string = "/token"
)

// GetBearerToken calls the /token SAS Viya Orders API endpoint to exchange client credentials for a Bearer token.
// The client credentials are obtained from the SAS API Portal (https://apiportal.sas.com), and should be defined in
// Viper (https://github.com/spf13/viper) as clientCredentialsId (key) and clientCredentialsSecret (secret).
func GetBearerToken() (token string, err error) {
	id, err := base64.StdEncoding.DecodeString(viper.GetString("clientCredentialsId"))
	if err != nil {
		return token, errors.New("ERROR: attempt to decode clientCredentialsId failed: " + err.Error())
	}
	sec, err := base64.StdEncoding.DecodeString(viper.GetString("clientCredentialsSecret"))
	if err != nil {
		return token, errors.New("ERROR: attempt to decode clientCredentialsSecret failed: " + err.Error())
	}

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

	oauthCfg := &clientcredentials.Config{
		ClientID:     string(id),
		ClientSecret: string(sec),
		TokenURL:     urlStr,
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}

	oaToken, err := oauthCfg.Token(context.Background())
	if err != nil {
		return token, errors.New("ERROR: Bearer token request failed: " + err.Error())
	}
	token = oaToken.AccessToken

	return token, nil
}
