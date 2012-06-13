package ym

import (
  "testing"
  // "github.com/ziutek/mymysql/mysql"
  // _ "github.com/ziutek/mymysql/native" // Native engine
  // "fmt"
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
	ManualClose("26b4c714d2879ac98543dbac2de76b8a")
}

func TestLineItemGetByInsertionOrder(*testing.T) {
  Verbose = true
  Open(cred)
  defer Close()
  
  LineItemServiceGetByInsertionOrder(1185043, 5000, 0)
  // if (len(token) < 10) {
    // panic("Invalid token. Not logged in")
  // } 
}
