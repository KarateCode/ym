package ym

import (
	"net/http"
	"encoding/xml"
	"io/ioutil"
	// "fmt"
	// "io"
	// "os"
)	

type lineItemService struct {
	url string
}
var LineItemService = lineItemService{ url:`line_item.php` }

type LineItem struct {
	Id               int    `xml:"id"`
	Description      string `xml:"description"`
	Comment          string `xml:"comment"`
	InsertionOrderId int    `xml:"insertion_order_id"`
	Active           bool   `xml:"active"`
	TargetProfileId  int    `xml:"target_profile_id"`
}
// -- figure out how to do ym.lineItemService.GetByInsertionOrder
// -- Parse xml results
// -- return proper values
// -- write structs for proprietary objects like line_item
func (service *lineItemService) GetByInsertionOrder(insertionOrderId, entriesOnPage, pageNum int) ([]LineItem, int) {
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
	
	req, err := http.NewRequest("POST", credentials.url + service.url, buffer)
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
	
	type GetByIO struct {
		XMLName xml.Name
		Body struct {
			XMLName  xml.Name
			Innerxml string "innerxml"
			Fault struct {
				XMLName     xml.Name `xml:"Fault"`
				Faultstring string   `xml:"faultstring"`
			} 
			GetByInsertionOrderResponse struct {
				XMLName     xml.Name `xml:"getByInsertionOrderResponse"`
				LineItems   struct {
					XMLName     xml.Name `xml:"line_items"`
					Items  []LineItem   `xml:"item"`
					// ReportToken string   `xml:"report_token"`
				}
			}
		}
	}
	
	getByIO := new(GetByIO)
	p, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil { 
		panic(readErr)
		// return nil, readErr 
	}
	
	errUnmarshall := xml.Unmarshal(p, getByIO)
	// if error != nil { return nil, readErr }
	if errUnmarshall != nil {
		panic(errUnmarshall)
	}
	
	// fmt.Printf("\nResponse: %v", getByIO)
	return getByIO.Body.GetByInsertionOrderResponse.LineItems.Items, len(getByIO.Body.GetByInsertionOrderResponse.LineItems.Items)
}