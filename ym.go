package ym

import (
	"fmt"
	"html"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"text/template"
	"bytes"
	"io"
	"os"
	"errors"
	"strings"
)

var token string
var Verbose bool = false
var credentials Auth
var templatePath string
var reportRoot string

type Auth struct {
  Login, Password, Env, url string
}

type ReportData []struct {
	COLUMN []string `xml:"COLUMN"`
}

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
	if err != nil { panic(err) }
	templatePath = strings.Split(os.Getenv("GOPATH"), ":")[0] + "/src/github.com/KarateCode/ym/"
	println(templatePath)
}

func Session(cred Auth, work func()) {
	Open(cred)
	defer Close()
	work()
}

func Open(cred Auth) error {
	credentials = cred
  
	if credentials.Env == "test" {
		credentials.url = "https://api-test.yieldmanager.com/api-1.33/"
		reportRoot = "https://api-test.yieldmanager.com/reports/"
	} else {
		credentials.url = "https://api.yieldmanager.com/api-1.33/"
		reportRoot = "https://api.yieldmanager.com/reports/"
	}
  
	tmpl, err := template.ParseFiles(templatePath + "templates/contact/login.xml")
	if err != nil { return err }
	
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, credentials)
	if err != nil { return err }
	
	req, err := http.NewRequest("POST", credentials.url + "contact.php", buffer)
	if err != nil { return err }
	
	res, error := http.DefaultClient.Do(req)  
	if error != nil { return error }
    
	type loginStruct struct {
		XMLName xml.Name
		Body struct {
			XMLName xml.Name
			Innerxml string "innerxml"
			Fault struct {
				XMLName xml.Name `xml:"Fault"`
				Faultstring string `xml:"faultstring"`
			} 
			LoginResponse struct {
				XMLName xml.Name `xml:"loginResponse"`
				Token string `xml:"token"`
			}
		}
	}
	loginObj := new(loginStruct)
	p, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil { return readErr }
	
	error = xml.Unmarshal(p, loginObj)
	if error != nil { return error }
    
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
	if err != nil { panic(err) }
	
	_, err = tmpl.Parse(bodyXml)
	if err != nil { panic(err) }
	
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, data)
	if err != nil { panic(err) }
	
	return buffer
}

// figure out how to do ym.lineItemService.GetByInsertionOrder
// Parse xml results
// return proper values
// possible write structs for proprietary objects like line_item
func LineItemServiceGetByInsertionOrder(insertionOrderId, entriesOnPage, pageNum int) ([]map[string]string, int) {
	type Data struct {
		Token string
		InsertionOrderId int
		EntriesOnPage int
		PageNum int
	}
	
	bodyXml := `{{define "Body"}}<n1:getByInsertionOrder env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:LineItemService">
		<token xsi:type="xsd:string">{{.Token}}</token>
		<insertion_order_id xsi:type="xsd:long">{{.InsertionOrderId}}</insertion_order_id>
		<entries_on_page xsi:type="xsd:long">{{.EntriesOnPage}}</entries_on_page>
		<page_num xsi:type="xsd:long">{{.PageNum}}</page_num>
		</n1:getByInsertionOrder>{{end}}`

    
	buffer := AssembleTemplate(bodyXml, Data{Token:token, InsertionOrderId:insertionOrderId, EntriesOnPage:entriesOnPage, PageNum:pageNum})
	
	req, err := http.NewRequest("POST", credentials.url + "line_item.php", buffer)
	if err != nil {
		println("error creating request")
		panic(err)
	}
	res, error := http.DefaultClient.Do(req)  
	if error != nil {
		println("error posting adhoc report")
		panic(error)
	}
	io.Copy(os.Stdout, res.Body)  
	return nil, 0
}

func ComplexReport(requestXml string) (*ReportData, error) {
	type Data struct {
		Token string
		XmlString string
	}
	
	bodyXml := `{{define "Body"}}<n1:requestViaXML env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:ReportService">
		<token xsi:type="xsd:string">{{.Token}}</token>
		<xml xsi:type="xsd:string">{{.XmlString}}</xml>
		</n1:requestViaXML>{{end}}`

	buffer := AssembleTemplate(bodyXml, Data{Token:token, XmlString:html.EscapeString(requestXml)})
	// io.Copy(os.Stdout, buffer)  
	req, err := http.NewRequest("POST", credentials.url + "report.php", buffer)
	if err != nil {
		panic(err)
	}
	res, error := http.DefaultClient.Do(req)  
	if error != nil {
		panic(error)
	}
	
	type RequestViaXml struct {
		XMLName xml.Name
		Body struct {
			XMLName xml.Name
			Innerxml string "innerxml"
			Fault struct {
				XMLName xml.Name `xml:"Fault"`
				Faultstring string `xml:"faultstring"`
			} 
			RequestViaXMLResponse struct {
				XMLName xml.Name `xml:"requestViaXMLResponse"`
				Token string `xml:"token"`
				ReportToken string `xml:"report_token"`
			}
		}
	}
	requestViaXml := new(RequestViaXml)
	p, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil { return nil, readErr }
	
	error = xml.Unmarshal(p, requestViaXml)
	if error != nil { return nil, readErr }
	
	// println(requestViaXml.Body.RequestViaXMLResponse.ReportToken)
	reportUrl, err := Status(requestViaXml.Body.RequestViaXMLResponse.ReportToken)
	println(reportUrl)
	if error != nil { return nil, readErr }
	
	downloadReq, downloadErr := http.NewRequest("GET", reportUrl, nil)
	if downloadErr != nil { return nil, downloadErr }
		
	downloadRes, error := http.DefaultClient.Do(downloadReq)  
	if error != nil {
		panic(error)
	}
	defer downloadRes.Body.Close()
	
	// io.Copy(os.Stdout, downloadRes.Body)  
	// return nil, nil
	
	type IoData struct {
		Response struct {
			XMLName xml.Name `xml:"RESPONSE"`
			Data struct {
				XMLName xml.Name `xml:"DATA"`
				RData ReportData `xml:"ROW"`
			}
		}
	}
	
	ioData := new(IoData)
	p, readErr = ioutil.ReadAll(downloadRes.Body)
	if readErr != nil { return nil, readErr }
	
	error = xml.Unmarshal(p, ioData)
	if error != nil { return nil, error }
	
	return &ioData.Response.Data.RData, nil
}

func Status(reportToken string) (string, error) {
	type StatusData struct {
		Token string
		ReportToken string
	}
	
	statusXml := `{{define "Body"}}<n1:status env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:ReportService">
		<token xsi:type="xsd:string">{{.Token}}</token>
		<report_token xsi:type="xsd:string">{{.ReportToken}}</report_token>
		</n1:status>{{end}}`
	
	buffer := AssembleTemplate(statusXml, StatusData{Token:token, ReportToken:reportToken})
	// io.Copy(os.Stdout, buffer)  
	req, err := http.NewRequest("POST", credentials.url + "report.php", buffer)
	if err != nil {
		println("error creating request")
		panic(err)
	}
	res, error := http.DefaultClient.Do(req)  
	if error != nil {
		println("error posting adhoc report")
		panic(error)
	}
	
	// io.Copy(os.Stdout, res.Body)  
	// return nil, 0
	
	type Status struct {
		XMLName xml.Name
		Body struct {
			XMLName xml.Name
			Innerxml string "innerxml"
			Fault struct {
				XMLName xml.Name `xml:"Fault"`
				Faultstring string `xml:"faultstring"`
			} 
			StatusResponse struct {
				XMLName xml.Name `xml:"statusResponse"`
				Token string `xml:"token"`
				UrlReport string `xml:"url_report"`
			}
		}
	}
	status := new(Status)
	p, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil { return "", readErr }
	
	error = xml.Unmarshal(p, status)
	if error != nil { return "", error }
	
	return status.Body.StatusResponse.UrlReport, nil
}

func Close() {
	ManualClose(token)
}
  
func ManualClose(manualToken string) {
	if credentials.Env == "test" {
		credentials.url = "https://api-test.yieldmanager.com/api-1.33/"
	} else {
		credentials.url = "https://api.yieldmanager.com/api-1.33/"
	}
	
	tmpl, err := template.ParseFiles(templatePath + "templates/contact/logout.xml")
	if err != nil { panic(err) }
	
	buffer := new(bytes.Buffer)
	err = tmpl.Execute(buffer, manualToken)
	if err != nil { panic(err) }
	
	req, reqErr := http.NewRequest("POST", credentials.url + "contact.php", buffer)
	if reqErr != nil { panic(reqErr) }
	
	res, resErr := http.DefaultClient.Do(req)  
	if resErr != nil { panic(resErr)  }
	if Verbose {
		println("\n** Logged out **")
		io.Copy(os.Stdout, res.Body)  
	}
}
