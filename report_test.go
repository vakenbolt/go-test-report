package main

import (
	"os"
	"testing"
)

func TestRead(t *testing.T) {
	t.Skip()
	var tmp templateData
	fi, _ := os.Open("test_report.html")
	err := tmp.readDataFromHTML(fi)
	if err != nil {
		t.Error(err)
	}
}
