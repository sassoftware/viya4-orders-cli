// Copyright © 2020, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
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
	"strings"
)

const checksumsFile string = "sas-bases/checksums.txt"
const viyaOrdersAPIHost string = "https://api.sas.com"
const viyaOrdersAPIBasePath string = "/mysas"
const viyaOrdersAPIOrdersPath string = "/orders"

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

func New(token string, assetName string, orderNum string, cadenceName string, cadenceVer string, filePath string,
	fileName string, outputFormat string) (ar AssetReq) {
	return AssetReq{
		token: token,
		aName: assetName,
		oNum:  orderNum,
		cName: cadenceName,
		cVer:  cadenceVer,
		fPath: filePath,
		fName: fileName,
		oFmt:  outputFormat,
	}
}

var output out
type out struct {
	OrderNumber  			string	`json:"orderNumber"`
	AssetName 				string	`json:"assetName"`
	AssetReqURL 			string	`json:"assetReqURL"`
	AssetLocation			string	`json:"assetLocation"`
	Cadence					string	`json:"cadence"`
	CadenceRelease  		string	`json:"cadenceRelease"`
}

func (ar AssetReq) GetAsset() error {
	// Make the API call to download the requested asset
	fileName, err := ar.makeReq()
	if err != nil {
		return err
	}

	// Set the output struct properties
	// Cadence is only applicable to deploymentAssets and license
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

// Determine the location where the asset will be saved on disk
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

// Print contents of the output struct in the format specified by the caller
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

// Build HTTP request
func (ar AssetReq) buildReq() (req *http.Request, err error) {
	reqURL, err := ar.buildURL()
	if err != nil {
		return req, err
	}
	output.AssetReqURL = reqURL

	var bearer = "Bearer " + ar.token
	req, err = http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return req, errors.New("ERROR: setup of asset request failed: " + err.Error())
	}
	req.Header.Add("Authorization", bearer)

	return req, nil
}

// Build request URL
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

// Make HTTP request and return the name of the file where the requested asset was saved
func (ar AssetReq) makeReq() (fileName string, err error) {
	req, err := ar.buildReq()
	if err != nil {
		return fileName, err
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fileName, errors.New("ERROR: asset request failed to complete: " + err.Error())
	}

	// Handle the response
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fileName, errors.New("ERROR: ioutil.ReadAll() returned: " + err.Error() +
				" on attempt to read response body from non-200 response code")
		}
		var em = "ERROR: asset request failed: "
		var emErr string
		if body != nil && len(body) > 0 {
			emErr = fmt.Sprintf("%s", string(body))
		} else {
			emErr = fmt.Sprintf("%d -- %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		}
		return fileName, errors.New(em+emErr)
	}

	// Determine where on disk we will save the asset
	fileName, err = ar.getFileName(resp.Header.Get("Content-Disposition"))
	if err != nil {
		return fileName, err
	}

	// Save asset to disk
	out, err := os.Create(fileName)
	if err != nil {
		return fileName, errors.New("ERROR: attempt to create output file " + fileName + " failed: " + err.Error())
	}
	defer out.Close()
	io.Copy(out, resp.Body)

	return fileName, nil
}

// Get the cadence name, version, and release if applicable
func (ar AssetReq) getCadenceInfo(file string) (string, string, error) {
	// Cadence release is only applicable to deployment assets
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
	for true {
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

	return "", "", errors.New("ERROR: unable to extract cadence from " + checksumsFile)
}

// Find and return the cadence information in the given byte array
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


