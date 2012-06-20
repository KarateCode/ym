package ym

import (
	"net/http"
	"encoding/xml"
	"io/ioutil"
	"errors"
	// "fmt"
	// "io"
	// "os"
)	

type targetProfileService struct {
	url string
}
var TargetProfileService = targetProfileService{ url:`target_profile.php` }

// type LineItem struct {
// 	Id               int    `xml:"id"`
// 	Description      string `xml:"description"`
// 	Comment          string `xml:"comment"`
// 	InsertionOrderId int    `xml:"insertion_order_id"`
// 	Active           bool   `xml:"active"`
// 	TargetProfileId  int    `xml:"target_profile_id"`
// }

func (service *targetProfileService) GetTargetSellerLineItems(ownerType string, ownerId int) (bool, []int) {
	type Data struct {
		Token string
		OwnerType string
		OwnerId int
	}
	
	bodyXml := `{{define "Body"}}<n1:getTargetSellerLineItems env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:LineItemService">
		<token xsi:type="xsd:string">{{.Token}}</token>
		<owner_type xsi:type="xsd:string">{{.OwnerType}}</owner_type>
		<owner_id xsi:type="xsd:long">{{.OwnerId}}</owner_id>
		</n1:getTargetSellerLineItems>{{end}}`

    
	buffer := AssembleTemplate(bodyXml, Data{Token:token, OwnerType:ownerType, OwnerId:ownerId})
	
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
	// return false, nil
	
	type GetByIO struct {
		XMLName xml.Name
		Body struct {
			XMLName  xml.Name
			Innerxml string "innerxml"
			Fault struct {
				XMLName     xml.Name `xml:"Fault"`
				Faultstring string   `xml:"faultstring"`
			}
			GetTargetSellerLineItemsResponse struct {
				XMLName     xml.Name `xml:"getTargetSellerLineItemsResponse"`
				
				SellerLineItemDefault bool `xml:"seller_line_item_default"`
				LineItems   struct {
					XMLName     xml.Name `xml:"line_items"`
					LineItemIds []int   `xml:"item"`
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
	
	// fmt.Printf("\nResponse: %+v", getByIO)
	// fmt.Printf("\nSellerLineItemDefault: '%v'\n", getByIO.Body.GetTargetSellerLineItemsResponse.SellerLineItemDefault)
	return getByIO.Body.GetTargetSellerLineItemsResponse.SellerLineItemDefault, getByIO.Body.GetTargetSellerLineItemsResponse.LineItems.LineItemIds
}
	
func (service *targetProfileService) SetTargetSellerLineItems(ownerType string, ownerId int, sellerLineItemDefault bool, lineItemIds []int, App bool) error {
	type Data struct {
		Token string
		OwnerType string
		OwnerId int
		SellerLineItemDefault bool
		LineItems []int
		Append bool
	}
	
	bodyXml := `{{define "Body"}}<n1:setTargetSellerLineItems env:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:n1="urn:LineItemService">
		<token xsi:type="xsd:string">{{.Token}}</token>
		<owner_type xmlns:n2="https://api.yieldmanager.com/types" xsi:type="n2:enum_target_profile_owner_type_ext">{{.OwnerType}}</owner_type>
		<owner_id xsi:type="xsd:long">{{.OwnerId}}</owner_id>
		<seller_line_item_default xsi:type="xsd:boolean">{{.SellerLineItemDefault}}</seller_line_item_default>
		<append xsi:type="xsd:boolean">{{.OwnerType}}</append>
		<line_items xmlns:n3="http://schemas.xmlsoap.org/soap/encoding/" n3:arrayType="xsd:long[1]" xsi:type="n3:Array">
		{{range .LineItems}}
			<item>{{.}}</item>
		{{end}}
		</line_items>
		</n1:setTargetSellerLineItems>{{end}}`

    
	buffer := AssembleTemplate(bodyXml, Data{Token:token, OwnerType:ownerType, OwnerId:ownerId, SellerLineItemDefault:sellerLineItemDefault, LineItems:lineItemIds, Append:App})
	
	req, err := http.NewRequest("POST", credentials.url + service.url, buffer)
	if err != nil {
		println("error creating request")
		panic(err)
		return err
	}
	res, error := http.DefaultClient.Do(req)  
	if error != nil {
		println("error posting adhoc report")
		panic(error)
		return error
	}
	// io.Copy(os.Stdout, res.Body)  
	// return nil
	
	type GetByIO struct {
		XMLName xml.Name
		Body struct {
			XMLName  xml.Name
			Innerxml string "innerxml"
			Fault struct {
				XMLName     xml.Name `xml:"Fault"`
				Faultstring string   `xml:"faultstring"`
			}
		}
	}
	
	getByIO := new(GetByIO)
	p, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil { 
		// panic(readErr)
		return readErr 
	}
	
	errUnmarshall := xml.Unmarshal(p, getByIO)
	// if error != nil { return nil, readErr }
	if errUnmarshall != nil {
		// panic(errUnmarshall)
		return errUnmarshall
	}
	
	if getByIO.Body.Fault.Faultstring != "" {
		return errors.New(getByIO.Body.Fault.Faultstring)
	}
	
	// fmt.Printf("\nResponse: %v", getByIO)
	return nil
}