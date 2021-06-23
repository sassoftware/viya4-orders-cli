// Copyright © 2020, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package assetreqs provides a method to request an order asset and receive printed information to STDOUT about it.
package assetreqs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

// checksumsFile is where we can find cadence information within downloaded deployment assets.
const checksumsFile string = "sas-bases/checksums.txt"

const (
	viyaOrdersAPIHost       string = "https://api.sas.com"
	viyaOrdersAPIBasePath   string = "/mysas"
	viyaOrdersAPIOrdersPath string = "/orders"
)

// AssetReq provides fields that define the parameters of an order asset request.
type AssetReq struct {
	token string
	aName string
	oNum  string
	cName string
	cVer  string
	fPath string
	fName string
	oFmt  string
}

// New initializes an AssetReq struct.
func New(token string, assetName string, orderNum string, cadenceName string, cadenceVer string, filePath string,
	fileName string, outputFormat string) (ar AssetReq, err error) {

	// Do some preliminary validation of the order number.
	valid, err := ordNumIsValidFmt(orderNum)
	if err != nil {
		return AssetReq{}, err
	}
	if !valid {
		return AssetReq{},
			errors.New("Given order number '" + orderNum + "' does not have the format of a valid order 	number.")
	}

	return AssetReq{
		token: token,
		aName: assetName,
		oNum:  orderNum,
		cName: cadenceName,
		cVer:  cadenceVer,
		fPath: filePath,
		fName: fileName,
		oFmt:  outputFormat,
	}, nil
}

func ordNumIsValidFmt(orderNum string) (v bool, err error) {
	// Note: because of no look-ahead operator (?) in go regexp, this takes multiple evaluations.
	// Make sure that the given order number has at least one number and at least one letter, contains only digits and
	// letters, and is 6 characters long.

	// Test length = 6.
	if len(orderNum) != 6 {
		return false, nil
	}

	// Length is correct - make sure it doesn't contain something other than letters or digits.
	invalid, err := hasNonLetterOrNonDigit(orderNum)
	if err != nil {
		return false, err
	}
	if invalid {
		// contains something other than letters and digits - invalid
		return false, err
	}

	// Make sure it contains at least one letter.
	v, err = hasLetter(orderNum)
	if err != nil {
		return false, err
	}
	if !v {
		// doesn't contain a letter
		return false, err
	}

	// Contains at least one letter - make sure it contains at least one digit.
	return hasDigit(orderNum)
}

func hasLetter(str string) (bool, error) {
	return regexp.Match("[a-zA-Z]+", []byte(str))
}

func hasDigit(str string) (bool, error) {
	return regexp.Match("[\\d]+", []byte(str))
}

func hasNonLetterOrNonDigit(str string) (bool, error) {
	return regexp.Match("[^a-zA-Z\\d]", []byte(str))
}

var output out

// out defines the information that is printed to STDOUT.
type out struct {
	OrderNumber    string `json:"orderNumber"`
	AssetName      string `json:"assetName"`
	AssetReqURL    string `json:"assetReqURL"`
	AssetLocation  string `json:"assetLocation"`
	Cadence        string `json:"cadence"`
	CadenceRelease string `json:"cadenceRelease"`
}

// GetAsset fetches the requested order asset (as defined in the AssetReq receiver) from the SAS Viya Orders API and
// prints information about it.
func (ar AssetReq) GetAsset() error {
	// Make the API call to download the requested asset
	fileName, err := ar.makeReq()
	if err != nil {
		return err
	}

	// Set the output struct properties.
	// Cadence is only applicable to deploymentAssets and license.
	if ar.aName == "deploymentAssets" || ar.aName == "license" {
		output.Cadence, output.CadenceRelease, err = ar.getCadenceInfo(fileName)
		if err != nil {
			return err
		}
	}
	output.OrderNumber = ar.oNum
	output.AssetName = ar.aName
	output.AssetLocation = fileName

	// Print the output
	err = ar.printOutput()
	if err != nil {
		return err
	}

	return nil
}

// getFileName determines the location where the asset will be saved on disk.
func (ar AssetReq) getFileName(contentDisp string) (fileName string, err error) {
	var filePath string
	if ar.fPath != "" {
		filePath = ar.fPath
	} else {
		filePath, err = os.Getwd()
		if err != nil {
			return fileName, errors.New("ERROR: os.Getwd() returned: " + err.Error())
		}
	}

	// Get the name of the file as returned by the API
	_, params, err := mime.ParseMediaType(contentDisp)
	if err != nil {
		return fileName, errors.New("ERROR: mime.ParseMediaType() returned: " + err.Error())
	}
	apiFNm := filepath.Join(filePath, params["filename"])

	if ar.fName != "" {
		// Even if they specified -n, use the extension that the API returned
		fileName = filepath.Join(filePath, ar.fName) + filepath.Ext(apiFNm)
	} else {
		fileName = apiFNm
	}

	return fileName, nil
}

// printOutput prints the contents of the output struct in the format specified by the caller.
func (ar AssetReq) printOutput() (err error) {
	if strings.ToLower(ar.oFmt) == "json" || strings.ToLower(ar.oFmt) == "j" {
		buff := new(bytes.Buffer)
		b, err := json.MarshalIndent(&output, "", "\t")
		if err != nil {
			return errors.New("ERROR: json.MarshalIndent() returned: " + err.Error())
		}
		buff.Write(b)
		_, err = buff.WriteTo(os.Stdout)
		if err != nil {
			return errors.New("ERROR: buff.WriteTo(os.Stdout) returned: " + err.Error())
		}
	} else {
		s := reflect.ValueOf(&output).Elem()
		typeOfT := s.Type()
		for i := 0; i < s.NumField(); i++ {
			f := s.Field(i)
			fmt.Printf("%s: %v\n",
				typeOfT.Field(i).Name, f.Interface())
		}
	}
	return nil
}

// buildReq builds an HTTP request.
func (ar AssetReq) buildReq() (req *http.Request, err error) {
	reqURL, err := ar.buildURL()
	if err != nil {
		return req, err
	}
	output.AssetReqURL = reqURL

	bearer := "Bearer " + ar.token
	req, err = http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return req, errors.New("ERROR: setup of asset request failed: " + err.Error())
	}
	req.Header.Set("Authorization", bearer)

	return req, nil
}

// buildURL builds the request URL.
func (ar AssetReq) buildURL() (urlStr string, err error) {
	u, err := url.ParseRequestURI(viyaOrdersAPIHost)
	if err != nil {
		return urlStr, errors.New("ERROR: attempt to parse asset request URI failed: " + err.Error())
	}

	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "%s", viyaOrdersAPIBasePath)
	_, _ = fmt.Fprintf(&b, "%s", viyaOrdersAPIOrdersPath)
	_, _ = fmt.Fprintf(&b, "%s", "/")
	_, _ = fmt.Fprintf(&b, "%s", ar.oNum)
	_, _ = fmt.Fprintf(&b, "%s", "/")

	if ar.cName != "" {
		_, _ = fmt.Fprintf(&b, "%s", "cadenceNames/")
		_, _ = fmt.Fprintf(&b, "%s", strings.ToLower(ar.cName))
		_, _ = fmt.Fprintf(&b, "%s", "/")
	}

	if ar.cVer != "" {
		_, _ = fmt.Fprintf(&b, "%s", "cadenceVersions/")
		_, _ = fmt.Fprintf(&b, "%s", strings.ToLower(ar.cVer))
		_, _ = fmt.Fprintf(&b, "%s", "/")
	}
	_, _ = fmt.Fprintf(&b, "%s", ar.aName)

	u.Path = b.String()
	urlStr = u.String()

	return urlStr, nil
}

// makeReq makes an HTTP request for an order asset and returns the name of the file where the requested asset was saved.
func (ar AssetReq) makeReq() (fileName string, err error) {
	req, err := ar.buildReq()
	if err != nil {
		return fileName, err
	}

	// Send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fileName, errors.New("ERROR: asset request failed to complete: " + err.Error())
	}

	// Handle the response.

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fileName, errors.New("ERROR: ioutil.ReadAll() returned: " + err.Error() +
				" on attempt to read response body from non-200 response code")
		}
		em := "ERROR: asset request failed: "
		var emErr string
		if len(body) > 0 {
			emErr = string(body)
		} else {
			emErr = fmt.Sprintf("%d -- %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return fileName, errors.New(em + emErr)
	}

	// Determine where on disk we will save the asset.
	fileName, err = ar.getFileName(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return fileName, err
	}

	// Save asset to disk.
	out, err := os.Create(fileName)
	if err != nil {
		return fileName, errors.New("ERROR: attempt to create output file " + fileName + " failed: " + err.Error())
	}

	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fileName, errors.New("ERROR: io.Copy() returned: " + err.Error() +
			" on attempt to write to " + fileName)
	}

	return fileName, nil
}

// getCadenceInfo gets the cadence name, version, and release, if applicable, for the retrieved order asset.
func (ar AssetReq) getCadenceInfo(file string) (string, string, error) {
	// Cadence release is only applicable to deployment assets.
	if ar.aName == "license" {
		return strings.Title(ar.cName) + " " + ar.cVer, "", nil
	}

	// Asset was deployment assets... Extract the cadence info from checksums.txt - this will
	// be helpful because it tells the caller the cadence version (which they may not have specified because they
	// just wanted the latest, but will need to know at some point) and the cadence release that they got.
	f, err := os.Open(file)
	if err != nil {
		return "", "", errors.New("ERROR: attempt to open " + file + " failed: " + err.Error())
	}

	defer f.Close()
	gzf, err := gzip.NewReader(f)
	if err != nil {
		return "", "", errors.New("ERROR: prepare to read " + file + " failed: " + err.Error())
	}

	tarReader := tar.NewReader(gzf)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return "", "", errors.New("ERROR: end of file reached in " + file + " before cadence information found")
		}
		if err != nil {
			return "", "", errors.New("ERROR: attempt to read " + file + " failed: " + err.Error())
		}

		if header.Name == checksumsFile {
			data := make([]byte, header.Size)
			_, err := tarReader.Read(data)
			if err != nil {
				return "", "", errors.New("ERROR: attempt to read " + checksumsFile + " failed: " + err.Error())
			}
			cVal, cRel := extractCadence(data)
			return cVal, cRel, nil
		}
	}
}

// extractCadence finds and returns the cadence information in the given byte array.
func extractCadence(data []byte) (string, string) {
	cLabelSt := bytes.Index(data, []byte("Cadence Display Name:"))
	tempData := data[cLabelSt:]
	fields := bytes.Fields(tempData)
	cValueSt := bytes.Index(tempData, fields[3])
	cValueEnd := bytes.Index(tempData, []byte("\n"))
	cValue := string(tempData[cValueSt:cValueEnd])

	cRelSt := bytes.Index(data, []byte("Cadence Release:"))
	tempData = data[cRelSt:]
	fields = bytes.Fields(tempData)
	cRel := string(fields[2])

	return cValue, cRel
}
