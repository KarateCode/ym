package ym

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	// "fmt"
	"io"
	"os"
	"time"
)

type lineItemService struct {
	url string
}

var LineItemService = lineItemService{url: `line_item.php`}

type LineItem struct {
	Id               int    `xml:"id"`
	Description      string `xml:"description"`
	Comment          string `xml:"comment"`
	InsertionOrderId int    `xml:"insertion_order_id"`
	Active           bool   `xml:"active"`
	TargetProfileId  string `xml:"target_profile_id"` // can't be an int, since the api sometimes hands back a nil and the parser panics.
}

type GetByIO struct {
	XMLName xml.Name
	Body    struct {
		XMLName  xml.Name
		Innerxml string "innerxml"
		Fault    struct {
			XMLName     xml.Name `xml:"Fault"`
			Faultstring string   `xml:"faultstring"`
		}
		GetByInsertionOrderResponse struct {
			XMLName   xml.Name `xml:"getByInsertionOrderResponse"`
			LineItems struct {
				XMLName xml.Name   `xml:"line_items"`
				Items   []LineItem `xml:"item"`
				// ReportToken string   `xml:"report_token"`
			}
		}
	}
}

func (service *lineItemService) GetByInsertionOrder(insertionOrderId, entriesOnPage, pageNum int) ([]LineItem, int) {
	type Data struct {
		Token            string
		InsertionOrderId int
		EntriesOnPage    int
		PageNum          int
	}

	bodyXml := `{{define "Body"}}<n1:getByInsertionOrder env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:LineItemService">
		<token xsi:type="xsd:string">{{.Token}}</token>
		<insertion_order_id xsi:type="xsd:long">{{.InsertionOrderId}}</insertion_order_id>
		<entries_on_page xsi:type="xsd:long">{{.EntriesOnPage}}</entries_on_page>
		<page_num xsi:type="xsd:long">{{.PageNum}}</page_num>
		</n1:getByInsertionOrder>{{end}}`

	buffer := AssembleTemplate(bodyXml, Data{Token: token, InsertionOrderId: insertionOrderId, EntriesOnPage: entriesOnPage, PageNum: pageNum})

	retries := 0
	getByIO := new(GetByIO)
	var res *http.Response
	var p []byte
	var readErr error
	for retries < 6 {
		req, err := http.NewRequest("POST", credentials.url+service.url, buffer)
		if err != nil {
			println("error creating request")
			panic(err)
		}
		res, err = http.DefaultClient.Do(req)
		if err != nil {
			println("error posting adhoc report")
			panic(err)
		}
		// io.Copy(os.Stdout, res.Body)
		// return nil, 0

		p, readErr = ioutil.ReadAll(res.Body)
		if readErr != nil {
			if retries > 6 {
				panic(readErr)
			}
			// return nil, readErr 
		} else {
			break
		}

		println("sleeping ", retries)
		time.Sleep(30 * time.Second)
		retries += 1
	}

	errUnmarshall := xml.Unmarshal(p, getByIO)
	// if error != nil { return nil, readErr }
	if errUnmarshall != nil {
		io.Copy(os.Stdout, res.Body)
		panic(errUnmarshall)
	}

	// fmt.Printf("\nResponse: %v", getByIO)
	return getByIO.Body.GetByInsertionOrderResponse.LineItems.Items, len(getByIO.Body.GetByInsertionOrderResponse.LineItems.Items)
}

func (service *lineItemService) GetByBuyer(insertionOrderId, entriesOnPage, pageNum int) ([]LineItem, int) {
	type Data struct {
		Token            string
		InsertionOrderId int
		EntriesOnPage    int
		PageNum          int
	}

	bodyXml := `{{define "Body"}}<n1:getByBuyer env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:LineItemService">
		<token xsi:type="xsd:string">{{.Token}}</token>
		<buyer_id xsi:type="xsd:long">{{.InsertionOrderId}}</buyer_id>
		<entries_on_page xsi:type="xsd:long">{{.EntriesOnPage}}</entries_on_page>
		<page_num xsi:type="xsd:long">{{.PageNum}}</page_num>
		</n1:getByBuyer>{{end}}`

	buffer := AssembleTemplate(bodyXml, Data{Token: token, InsertionOrderId: insertionOrderId, EntriesOnPage: entriesOnPage, PageNum: pageNum})

	req, err := http.NewRequest("POST", credentials.url+service.url, buffer)
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

	type GetByBuyer struct {
		XMLName xml.Name
		Body    struct {
			XMLName  xml.Name
			Innerxml string "innerxml"
			Fault    struct {
				XMLName     xml.Name `xml:"Fault"`
				Faultstring string   `xml:"faultstring"`
			}
			GetByInsertionOrderResponse struct {
				XMLName   xml.Name `xml:"getByBuyerResponse"`
				LineItems struct {
					XMLName xml.Name   `xml:"line_items"`
					Items   []LineItem `xml:"item"`
					// ReportToken string   `xml:"report_token"`
				}
			}
		}
	}

	getByBuyer := new(GetByBuyer)
	p, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		panic(readErr)
		// return nil, readErr 
	}

	errUnmarshall := xml.Unmarshal(p, getByBuyer)
	// if error != nil { return nil, readErr }
	if errUnmarshall != nil {
		io.Copy(os.Stdout, res.Body)
		panic(errUnmarshall)
	}

	// fmt.Printf("\nResponse: %v", getByIO)
	return getByBuyer.Body.GetByInsertionOrderResponse.LineItems.Items, len(getByBuyer.Body.GetByInsertionOrderResponse.LineItems.Items)
}
