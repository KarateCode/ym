package ym

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

type Config struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
}

func TimeoutDialer(config *Config) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, config.ConnectTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(config.ReadWriteTimeout))
		return conn, nil
	}
}

func NewTimeoutClient(args ...interface{}) *http.Client {
	// Default configuration
	config := &Config{
		ConnectTimeout:   1 * time.Second,
		ReadWriteTimeout: 1 * time.Second,
	}

	// merge the default with user input if there is one
	if len(args) == 1 {
		timeout := args[0].(time.Duration)
		config.ConnectTimeout = timeout
		config.ReadWriteTimeout = timeout
	}

	if len(args) == 2 {
		config.ConnectTimeout = args[0].(time.Duration)
		config.ReadWriteTimeout = args[1].(time.Duration)
	}

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(config),
		},
	}
}

var token string
var Verbose bool = false
var credentials Auth
var templatePath string
var reportRoot string
var version = "1.36"

type Auth struct {
	Login, Password, Env, url string
}

type Column []string
type Row struct {
	Column Column `xml:"COLUMN"`
}
type ReportData []Row

var layoutXml = `<?xml version="1.0" encoding="utf-8" ?>
<env:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:xsd="http://www.w3.org/2001/XMLSchema"
    xmlns:env="http://schemas.xmlsoap.org/soap/envelope/">
  <env:Body>
    {{template "Body" .}}
  </env:Body>
</env:Envelope>`
var layoutTemplate *template.Template

func init() {
	var err error
	layoutTemplate, err = template.New("Layout").Parse(layoutXml)
	if err != nil {
		panic(err)
	}
	templatePath = strings.Split(os.Getenv("GOPATH"), ":")[0] + "/src/github.com/KarateCode/ym/"
}

func Session(cred Auth, work func()) {
	Open(cred)
	defer Close()
	work()
}

func Open(cred Auth) error {
	credentials = cred

	if credentials.Env == "test" {
		credentials.url = "https://api-test.yieldmanager.com/api-" + version + "/"
		reportRoot = "https://api-test.yieldmanager.com/reports/"
	} else {
		credentials.url = "https://api.yieldmanager.com/api-" + version + "/"
		reportRoot = "https://api.yieldmanager.com/reports/"
	}

	tmpl, err := template.ParseFiles(templatePath + "templates/contact/login.xml")
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, credentials)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", credentials.url+"contact.php", buffer)
	if err != nil {
		return err
	}

	res, error := http.DefaultClient.Do(req)
	if error != nil {
		return error
	}

	type loginStruct struct {
		XMLName xml.Name
		Body    struct {
			XMLName  xml.Name
			Innerxml string "innerxml"
			Fault    struct {
				XMLName     xml.Name `xml:"Fault"`
				Faultstring string   `xml:"faultstring"`
			}
			LoginResponse struct {
				XMLName xml.Name `xml:"loginResponse"`
				Token   string   `xml:"token"`
			}
		}
	}
	loginObj := new(loginStruct)
	p, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return readErr
	}

	error = xml.Unmarshal(p, loginObj)
	if error != nil {
		return error
	}

	token = loginObj.Body.LoginResponse.Token
	if Verbose {
		println("\n** Logged in:", credentials.url, " **")
		fmt.Printf("%s\n", p)
	}
	if len(loginObj.Body.Fault.Faultstring) > 0 {
		return errors.New(loginObj.Body.Fault.Faultstring)
	}

	return nil
}

func AssembleTemplate(bodyXml string, data interface{}) *bytes.Buffer {
	tmpl, err := layoutTemplate.Clone()
	if err != nil {
		panic(err)
	}

	_, err = tmpl.Parse(bodyXml)
	if err != nil {
		panic(err)
	}

	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, data)
	if err != nil {
		panic(err)
	}

	return buffer
}

func checkConnection() {
	if token == "" {
		panic(errors.New("No valid token. Did you forget to open a session?"))
	}
}

type IoData struct {
	Response struct {
		XMLName xml.Name `xml:"RESPONSE"`
		Data    struct {
			XMLName xml.Name `xml:"DATA"`

			Header Row        `xml:"HEADER"`
			RData  ReportData `xml:"ROW"`
		}
	}
}

func ComplexReport(requestXml string) (*ReportData, *Row, error) {
	checkConnection()

	type Data struct {
		Token     string
		XmlString string
	}

	bodyXml := `{{define "Body"}}<n1:requestViaXML env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:ReportService">
		<token xsi:type="xsd:string">{{.Token}}</token>
		<xml xsi:type="xsd:string">{{.XmlString}}</xml>
		</n1:requestViaXML>{{end}}`

	buffer := AssembleTemplate(bodyXml, Data{Token: token, XmlString: html.EscapeString(requestXml)})
	// io.Copy(os.Stdout, buffer)
	// println(credentials.url + "report.php")
	req, err := http.NewRequest("POST", credentials.url+"report.php", buffer)
	if err != nil {
		panic(err)
	}
	// client := http.DefaultClient

	retries := 0
	var res *http.Response
	for retries < 20 {
		client := NewTimeoutClient(500*time.Millisecond, 10*time.Minute)
		res, err = client.Do(req)
		if err != nil {
			fmt.Printf("err: %+v\n", err)
			if retries >= 6 {
				panic(err)
			}
		} else {
			break
		}

		// reportUrl, err = Status(requestViaXml.Body.RequestViaXMLResponse.ReportToken)
		// if err != nil {
		// 	panic(err)
		// 	// return nil, readErr
		// }
		// if reportUrl != "" && reportUrl != "https://api-test.yieldmanager.com/reports/" {
		// 	break
		// }
		// println("reportUrl:", reportUrl)
		// println("sleeping ", retries)
		time.Sleep(30 * time.Second)
		// println("reattempting ", requestViaXml.Body.RequestViaXMLResponse.ReportToken)
		retries += 1
	}

	// io.Copy(os.Stdout, res.Body)

	type RequestViaXml struct {
		XMLName xml.Name
		Body    struct {
			XMLName  xml.Name
			Innerxml string "innerxml"
			Fault    struct {
				XMLName     xml.Name `xml:"Fault"`
				Faultstring string   `xml:"faultstring"`
			}
			RequestViaXMLResponse struct {
				XMLName     xml.Name `xml:"requestViaXMLResponse"`
				Token       string   `xml:"token"`
				ReportToken string   `xml:"report_token"`
			}
		}
	}
	requestViaXml := new(RequestViaXml)
	p, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		panic(readErr)
		// return nil, readErr 
	}

	errUnmarshall := xml.Unmarshal(p, requestViaXml)
	// if error != nil { return nil, readErr }
	if errUnmarshall != nil {
		panic(errUnmarshall)
	}
	if len(requestViaXml.Body.Fault.Faultstring) > 0 {
		println(requestViaXml.Body.Fault.Faultstring)
		return nil, nil, errors.New(requestViaXml.Body.Fault.Faultstring)
	}

	// println(requestViaXml.Body.RequestViaXMLResponse.ReportToken)
	retries = 0
	var reportUrl string
	for retries < 20 {
		reportUrl, err = Status(requestViaXml.Body.RequestViaXMLResponse.ReportToken)
		if err != nil {
			// panic(err)
			fmt.Printf("err: %+v\n", err)
			// return nil, readErr
			if strings.Contains(err.Error(), `While beginning to sink the request, an exception occurred. Message: The number of dynamic conversions (advertiser.convs) has exceeded maximum limit of 10000. Please narrow your request filters.' occurred while report system was fulfilling your request.`) {
				return nil, nil, err
			}
		} else if reportUrl != "" && reportUrl != "https://api-test.yieldmanager.com/reports/" {
			break
		}
		println("reportUrl:", reportUrl)
		println("sleeping ", retries)
		time.Sleep(30 * time.Second)
		println("reattempting status", requestViaXml.Body.RequestViaXMLResponse.ReportToken)
		retries += 1
	}

	downloadReq, downloadErr := http.NewRequest("GET", reportUrl, nil)
	if downloadErr != nil {
		// return nil, downloadErr 
		panic(downloadErr)
	}

	retries = 0
	ioData := new(IoData)
	for retries < 6 {
		client := NewTimeoutClient(500*time.Millisecond, 10*time.Minute)
		// client := http.DefaultClient
		downloadRes, errGet := client.Do(downloadReq)
		if errGet != nil {
			if retries > 6 {
				panic(errGet)
			}
			fmt.Printf("errGet: %+v\n", errGet)
			time.Sleep(15 * time.Second)
			println("reattempting download ", requestViaXml.Body.RequestViaXMLResponse.ReportToken)
			retries += 1
		} else {
			defer downloadRes.Body.Close()
			// io.Copy(os.Stdout, downloadRes.Body)
			// return nil, nil

			p, readErr = ioutil.ReadAll(downloadRes.Body)
			if readErr != nil {
				// return nil, readErr 
				panic(readErr)
			}
			// var xmlBody string = string(p)
			// if bytes.IndexRune(p, '\u0005') > -1 {
			// 	println(len(p))
			// 	println("found match")
			// 	println(bytes.IndexRune(p, '\u0005'))
			// }
			// errUnmarshall = xml.Unmarshal(bytes.Replace(p, '\u0005', []byte("")), ioData)
			errUnmarshall = xml.Unmarshal(bytes.Map(mapOutIllChars, p), ioData)
			// errUnmarshall = xml.Unmarshal(p, ioData)
			if errUnmarshall != nil {
				// return nil, errUnmarshall 
				println("\n", reportUrl, "\n")
				println(string(p))
				println("sleeping")
				println(errUnmarshall.Error())
				time.Sleep(15 * time.Second)
				println("reattempting")
				retries += 1
				if retries >= 6 {
					// panic(errUnmarshall)
					log.Print("Retries exceeded. Skipping ", requestXml)
				}
			} else {
				break
			}
		}
	}

	// fmt.Printf("Header: %+v\n", ioData.Response.Data)
	return &ioData.Response.Data.RData, &ioData.Response.Data.Header, nil
}

func mapOutIllChars(r rune) rune {
	if r == '\u0005' {
		return -1
	}
	return r
}

func Status(reportToken string) (string, error) {
	type StatusData struct {
		Token       string
		ReportToken string
	}

	statusXml := `{{define "Body"}}<n1:status env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:ReportService">
		<token xsi:type="xsd:string">{{.Token}}</token>
		<report_token xsi:type="xsd:string">{{.ReportToken}}</report_token>
		</n1:status>{{end}}`

	buffer := AssembleTemplate(statusXml, StatusData{Token: token, ReportToken: reportToken})
	// io.Copy(os.Stdout, buffer)  
	req, err := http.NewRequest("POST", credentials.url+"report.php", buffer)
	if err != nil {
		println("error creating request")
		panic(err)
	}
	client := NewTimeoutClient(500*time.Millisecond, 10*time.Minute)
	// client := http.DefaultClient
	res, err := client.Do(req)
	if err != nil {
		println("error posting adhoc report")
		// panic(error)
		return "", err
	}

	// io.Copy(os.Stdout, res.Body)  
	// return "", nil

	type Status struct {
		XMLName xml.Name
		Body    struct {
			XMLName  xml.Name
			Innerxml string "innerxml"
			Fault    struct {
				XMLName     xml.Name `xml:"Fault"`
				Faultstring string   `xml:"faultstring"`
			}
			StatusResponse struct {
				XMLName   xml.Name `xml:"statusResponse"`
				Token     string   `xml:"token"`
				UrlReport string   `xml:"url_report"`
			}
		}
	}
	status := new(Status)
	p, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return "", readErr
	}

	err = xml.Unmarshal(p, status)
	if err != nil {
		return "", err
	}

	if len(status.Body.Fault.Faultstring) > 0 {
		return "", errors.New(status.Body.Fault.Faultstring)
	}

	return status.Body.StatusResponse.UrlReport, nil
}

func Close() {
	ManualClose(token)
}

func ManualClose(manualToken string) {
	if credentials.Env == "test" {
		credentials.url = "https://api-test.yieldmanager.com/api-" + version + "/"
	} else {
		credentials.url = "https://api.yieldmanager.com/api-" + version + "/"
	}

	tmpl, err := template.ParseFiles(templatePath + "templates/contact/logout.xml")
	if err != nil {
		panic(err)
	}

	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, manualToken)
	if err != nil {
		panic(err)
	}

	req, reqErr := http.NewRequest("POST", credentials.url+"contact.php", buffer)
	if reqErr != nil {
		panic(reqErr)
	}

	res, resErr := http.DefaultClient.Do(req)
	if resErr != nil {
		panic(resErr)
	}
	if Verbose {
		println("\n** Logged out **")
		io.Copy(os.Stdout, res.Body)
	}
	token = ""
}
