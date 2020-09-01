//Copyright © 2020, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
//SPDX-License-Identifier: Apache-2.0
package assetreqs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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

func (ar AssetReq) GetAsset() {
	//Make the API call to download the requested asset
	fileName := ar.makeReq()

	//Set the output struct properties
	//Cadence is only applicable to deploymentAssets and license
	if ar.aName == "deploymentAssets" || ar.aName == "license" {
		output.Cadence, output.CadenceRelease = ar.getCadenceInfo(fileName)
	}
	output.OrderNumber = ar.oNum
	output.AssetName = ar.aName
	output.AssetLocation = fileName

	//Print the output
	ar.printOutput()
}

//Determine the location where the asset will be saved on disk
func (ar AssetReq) getFileName(contentDisp string) string {
	var filePath string
	var err error
	if ar.fPath != "" {
		filePath = ar.fPath
	} else {
		filePath, err = os.Getwd()
		if err != nil {
			log.Panic("Error attempting to get current working directory: " + err.Error())
		}
	}

	//Get the name of the file as returned by the API
	_, params, err := mime.ParseMediaType(contentDisp)
	if err != nil {
		log.Panic("Error attempting to parse Content-Disposition header in API response: " + err.Error())
	}
	apiFNm := filepath.Join(filePath, params["filename"])

	var fileName string
	if ar.fName != "" {
		//Even if they specified -n, use the extension that the API returned
		fileName = filepath.Join(filePath, ar.fName) + filepath.Ext(apiFNm)
	} else {
		fileName = apiFNm
	}

	return fileName
}

//Print contents of the output struct in the format specified by the caller
func (ar AssetReq) printOutput() {
	if strings.ToLower(ar.oFmt) == "json" || strings.ToLower(ar.oFmt) == "j" {
		buff := new(bytes.Buffer)
		b, err := json.MarshalIndent(&output, "", "\t")
		if err != nil {
			log.Panic("json.MarshalIndent() returned: " + err.Error())
		}
		buff.Write(b)
		buff.WriteTo(os.Stdout)
	} else {
		s := reflect.ValueOf(&output).Elem()
		typeOfT := s.Type()
		for i := 0; i < s.NumField(); i++ {
			f := s.Field(i)
			fmt.Printf("%s: %v\n",
				typeOfT.Field(i).Name, f.Interface())
		}
	}
}

//Build HTTP request
func (ar AssetReq) buildReq() *http.Request {
	reqURL := ar.buildURL()
	output.AssetReqURL = reqURL

	var bearer = "Bearer " + ar.token
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Panic("Error setting up asset request: " + err.Error())
	}
	req.Header.Add("Authorization", bearer)

	return req
}

//Build request URL
func (ar AssetReq) buildURL() string {
	u, err := url.ParseRequestURI(viyaOrdersAPIHost)
	if err != nil {
		log.Panic("Error parsing asset request URI: " + err.Error())
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s", viyaOrdersAPIBasePath)
	fmt.Fprintf(&b, "%s", viyaOrdersAPIOrdersPath)
	fmt.Fprintf(&b, "%s", "/")
	fmt.Fprintf(&b, "%s", ar.oNum)
	fmt.Fprintf(&b, "%s", "/")

	if ar.cName != "" {
		fmt.Fprintf(&b, "%s", "cadenceNames/")
		fmt.Fprintf(&b, "%s", strings.ToLower(ar.cName))
		fmt.Fprintf(&b, "%s", "/")
	}

	if ar.cVer != "" {
		fmt.Fprintf(&b, "%s", "cadenceVersions/")
		fmt.Fprintf(&b, "%s", strings.ToLower(ar.cVer))
		fmt.Fprintf(&b, "%s", "/")
	}
	fmt.Fprintf(&b, "%s", ar.aName)

	u.Path = b.String()

	return u.String()
}

//Make HTTP request and return the name of the file where the requested asset was saved
func (ar AssetReq) makeReq() string {
	req := ar.buildReq()

	//Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic("Error on asset request: ", err.Error())
	}

	//Handle the response
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Panic("Error reading error response body from asset request: " + err.Error())
		}
		var em = fmt.Sprintf("Asset request returned: %d -- %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		if body != nil && len(body) > 0 {
			em += fmt.Sprintf(". Error: %s", string(body))
		}
		log.Panic(em)
	}

	//Determine where on disk we will save the asset
	fileName := ar.getFileName(resp.Header.Get("Content-Disposition"))

	//Save asset to disk
	out, err := os.Create(fileName)
	if err != nil {
		log.Panic("Error attempting to create output file " + fileName + ": " + err.Error())
	}
	defer out.Close()
	io.Copy(out, resp.Body)


	return fileName
}

//Get the cadence name, version, and release if applicable
func (ar AssetReq) getCadenceInfo(file string) (string, string) {
	//Cadence release is only applicable to deployment assets
	if ar.aName == "license" {
		return strings.ToTitle(ar.cName) + " " + ar.cVer, ""
	}

	//Asset was deployment assets... Extract the cadence info from checksums.txt - this will
	//be helpful because it tells the caller the cadence version (which they may not have specified because they
	//just wanted the latest, but will need to know at some point) and the cadence release that they got.
	f, err := os.Open(file)
	if err != nil {
		log.Panic("Error attempting to open " + file + " : " + err.Error())
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		log.Panic("Error preparing to read " + file + ": " + err.Error())
	}

	tarReader := tar.NewReader(gzf)
	for true {
		header, err := tarReader.Next()
		if err == io.EOF {
			log.Panic("End of file reached in " + file + " before cadence information found")
			break
		}
		if err != nil {
			log.Panic("Error attempting to read " + file + " :" + err.Error())
		}

		if header.Name == checksumsFile {

			data := make([]byte, header.Size)
			_, err := tarReader.Read(data)
			if err != nil {
				log.Panic("Error reading " + checksumsFile + " : " + err.Error())
			}
			return extractCadence(data)
		}
	}

	log.Panic("Unable to extract cadence from " + checksumsFile)
	return "", ""
}

//Find and return the cadence information in the given byte array
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


