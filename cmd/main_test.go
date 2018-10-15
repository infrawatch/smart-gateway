package main

import (
	"fmt"
	"os"
	"testing"
)

func TestServiceTypeExists(t *testing.T) {
	var serviceType string
	os.Args = []string{"cmd", "-servicetype=bla"}
	serviceType = getServiceType()
	fmt.Println(serviceType)
	if serviceType != "bla" {
		t.Error("failed : command line argument 'servicetype' was not found")
	}

}

func TestServiceTypeDoesntExists(t *testing.T) {
	var serviceType string
	os.Args = []string{"cmd", "-servicenottype="}
	serviceType = getServiceType()
	//if valid not passed it will return defualt
	if serviceType == "" {
		t.Error("Failed for known command line argument 'servicenottype'.")
	}
}
