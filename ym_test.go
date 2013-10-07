package ym

import (
	"testing"
	// "github.com/ziutek/mymysql/mysql"
	// _ "github.com/ziutek/mymysql/native" // Native engine
	"encoding/xml"
	"fmt"
	// "strconv"
	// "os"
	// "time"
)

var cred = Auth{Login: "???", Password: "???", Env: "test"}

// func TestOpen(*testing.T) {
// Verbose = true
// Open(cred)
// defer Close()
// if (len(token) < 10) {
// panic("Invalid token. Not logged in")
// } 
// }

// func TestOpenErrorWrongPassword(*testing.T) {
// Verbose = true
// // token = "d67a45c886007ac4c4228690a12eba68"
// // Close()
// error := Open(Auth{Login: "???", Password: "???", Env: "test"})
// if (error == nil) {
// panic("Open should return an error")
// }
// // println(error.Error())
// }

// func TestSession(*testing.T) {
// Session(cred, func() {
// if (len(token) < 10) {
// panic("Invalid token. Not logged in")
// } 
// })
// }

func TestManualClose(*testing.T) {
	Verbose = true
	ManualClose("2ba2d298f264393ec976d6e94944133d")
}

func TestLineItemGetByInsertionOrder(*testing.T) {
	Verbose = true
	Open(cred)
	defer Close()

	item, _ := LineItemService.GetByInsertionOrder(10596, 5000, 0)
	fmt.Printf("item: %+v\n", item)
	// if (len(token) < 10) {
	// panic("Invalid token. Not logged in")
	// } 
}

func TestXmlParsingError(*testing.T) {
	var p []byte = []byte(`<?xml version="1.0" ?><RWResponse><RESPONSE><DATA><HEADER><COLUMN>interval</COLUMN><COLUMN>advertiser_id</COLUMN><COLUMN>advertiser_name</COLUMN><COLUMN>advertiser_io_name</COLUMN><COLUMN>advertiser_io_id</COLUMN><COLUMN>advertiser_line_item_name</COLUMN><COLUMN>advertiser_line_item_id</COLUMN><COLUMN>advertiser_entity_type</COLUMN><COLUMN>seller_imps</COLUMN><COLUMN>seller_clicks</COLUMN><COLUMN>seller_convs</COLUMN><COLUMN>network_gross_revenue</COLUMN><COLUMN>network_gross_cost</COLUMN></HEADER>
<ROW><COLUMN data_type="text">Oct 03, 2013</COLUMN><COLUMN data_type="numeric">388988</COLUMN><COLUMN data_type="text" id="388988">Bill Hudson and Associates, Inc.</COLUMN><COLUMN data_type="text" id="1687294">IU BallIO</COLUMN><COLUMN data_type="numeric">1687294</COLUMN><COLUMN data_type="text" id="8238883">IU Ball</COLUMN><COLUMN data_type="numeric">8238883</COLUMN><COLUMN data_type="text" id="388988">Managed Advertiser</COLUMN><COLUMN data_type="numeric">62999</COLUMN><COLUMN data_type="numeric">117</COLUMN><COLUMN data_type="numeric">0</COLUMN><COLUMN data_type="money" currency_type="n_prfrd" currency_id="153" currency_abbr="USD">114.051521</COLUMN><COLUMN data_type="money" currency_type="n_prfrd" currency_id="153" currency_abbr="USD">69.821236</COLUMN></ROW>
<ROW><COLUMN data_type="text">Oct 04, 2013</COLUMN><COLUMN data_type="numeric">388988</COLUMN><COLUMN data_type="text" id="388988">Bill Hudson and Associates, Inc.</COLUMN><COLUMN data_type="text" id="1687294">IU BallIO</COLUMN><COLUMN data_type="numeric">1687294</COLUMN><COLUMN data_type="text" id="8238883">IU Ball</COLUMN><COLUMN data_type="numeric">8238883</COLUMN><COLUMN data_type="text" id="388988">Managed Advertiser</COLUMN><COLUMN data_type="numeric">56580</COLUMN><COLUMN data_type="numeric">134</COLUMN><COLUMN data_type="numeric">0</COLUMN><COLUMN data_type="money" currency_type="n_prfrd" currency_id="153" currency_abbr="USD">109.309213</COLUMN><COLUMN data_type="money" currency_type="n_prfrd" currency_id="153" currency_abbr="USD">64.487018</COLUMN></ROW>
</DATA><METADATA transactionId="baa1f664-628a-3c8d-0ace-3e0400000897" tracking_id="2219166" rows="2" columns="13" domain="network" context="rpt" currency_type="NETWORK_PRFRD" timestart="Oct 03, 2013 00:00" timeend="Oct 07, 2013 00:00" granularity="daily_local" timezone="America/New_York" runat="Oct 07, 2013 09:59"></METADATA></RESPONSE></RWResponse>
<!-- sarq11.ngd.ne1.yahoo.com uncompressed/chunked Mon Oct  7 13:59:24 UTC 2013 -->`)
	ioData := new(IoData)

	errUnmarshall := xml.Unmarshal(p, ioData)
	if errUnmarshall != nil {
		// return nil, errUnmarshall 
		println(string(p))
		println("sleeping")
		println(errUnmarshall.Error())
		// time.Sleep(15 * time.Second)
		println("reattempting")
	}
	fmt.Printf("ioData: %+v\n", ioData)
}
