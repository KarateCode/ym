package main

import (
  "fmt"
  "encoding/xml"
  "io/ioutil"
  "net/http"
  "text/template"
  "bytes"
  "io"
  "os"
  "errors"
)

var token string
var Verbose bool = false
var credentials Auth

type Auth struct {
  Login, Password, Env, url string
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
}

func Session(cred Auth, work func()) {
  Open(cred)
  defer Close()
  work()
}

func Open(cred Auth) error {
  credentials = cred
  
  if credentials.Env == "test" {
    credentials.url = "https://api-test.yieldmanager.com/api-1.33/contact.php"
  } else {
    credentials.url = "https://api.yieldmanager.com/api-1.33/contact.php"
  }
  
  tmpl, err := template.ParseFiles("templates/contact/login.xml")
  if err != nil { return err }
  
  buffer := new(bytes.Buffer)
  err = tmpl.Execute(buffer, credentials)
  if err != nil { return err }
  
  req, err := http.NewRequest("POST", credentials.url, buffer)
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
    println("\n** Logged in **")
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
  bodyXml := `{{define "Body"}}<n1:getByInsertionOrder env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:LineItemService">
      <token xsi:type="xsd:string">{{.Token}}</token>
      <insertion_order_id xsi:type="xsd:long">{{.InsertionOrderId}}</insertion_order_id>
      <entries_on_page xsi:type="xsd:long">{{.EntriesOnPage}}</entries_on_page>
      <page_num xsi:type="xsd:long">{{.PageNum}}</page_num>
    </n1:getByInsertionOrder>{{end}}`

  type Data struct {
    Token string
    InsertionOrderId int
    EntriesOnPage int
    PageNum int
  }
    
  buffer := AssembleTemplate(bodyXml, Data{Token:token, InsertionOrderId:insertionOrderId, EntriesOnPage:entriesOnPage, PageNum:pageNum})

  req, err := http.NewRequest("POST", "https://api-test.yieldmanager.com/api-1.33/line_item.php", buffer)
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

func Close() {
  tmpl, err := template.ParseFiles("templates/contact/logout.xml")
  if err != nil { panic(err) }
    
  buffer := new(bytes.Buffer)
  err = tmpl.Execute(buffer, token)
  if err != nil { panic(err) }

  req, reqErr := http.NewRequest("POST", credentials.url, buffer)
  if reqErr != nil { panic(reqErr) }
    
  res, resErr := http.DefaultClient.Do(req)  
  if resErr != nil { panic(resErr)  }
  if Verbose {
    println("** Logged out **")
    io.Copy(os.Stdout, res.Body)  
  }
}
  
