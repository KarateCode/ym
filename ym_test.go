package ym

import (
	"testing"
	// "github.com/ziutek/mymysql/mysql"
	// _ "github.com/ziutek/mymysql/native" // Native engine
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
	ManualClose("57e3dd641dbbb0e11f42b886acd8fc7d")
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
