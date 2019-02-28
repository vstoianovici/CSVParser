package main

import (
	"fmt"
	"testing"
)

type MockEnvVar struct {
	files        string
	buildsExec   bool
	topUsers     bool
	successRate  bool
	topExitCodes bool
	startTime    string
	endTime      string
	sinceTime    string
	last         int64
	unit         string
	result       bool
}

var mockEnvVars = []MockEnvVar{

	{"./testdata/1_aFewValues.csv", true, true, true, true, "10/31/2018", "12/12/2018", "", 0, "", true},
	{"./testdata/2_exitCodes.csv", false, false, false, true, "10/31/2018", "12/12/2018", "", 0, "", true},
	{"./testdata/3_outOfOrder.csv", false, true, true, true, "10/31/2018", "12/12/2018", "", 0, "", true},
	{"./testdata/4_emptyLines.csv", false, true, true, true, "10/31/2018", "12/12/2018", "", 0, "", true},
	{"./testdata/1_aFewValues.csv", true, true, true, true, "", "", "", 6, "months", true},
	{"./testdata/2_exitCodes.csv", false, false, false, true, "", "", "", 6, "months", true},
	{"./testdata/3_outOfOrder.csv", false, true, true, true, "", "", "", 6, "months", true},
	{"./testdata/4_emptyLines.csv", false, true, true, true, "", "", "", 6, "months", true},
	{"./testdata/1_aFewValues.csv", true, true, true, true, "", "", "", 25, "weeks", true},
	{"./testdata/2_exitCodes.csv", false, false, false, true, "", "", "", 25, "weeks", true},
	{"./testdata/3_outOfOrder.csv", false, true, true, true, "", "", "", 25, "weeks", true},
	{"./testdata/4_emptyLines.csv", false, true, true, true, "", "", "", 25, "weeks", true},
	{"./testdata/1_aFewValues.csv", true, true, true, true, "", "", "10/31/2018", 0, "", true},
	{"./testdata/2_exitCodes.csv", false, false, false, true, "", "", "10/31/2018", 0, "", true},
	{"./testdata/3_outOfOrder.csv", false, true, true, true, "", "", "10/31/2018", 0, "", true},
	{"./testdata/4_emptyLines.csv", false, true, true, true, "", "", "10/31/2018", 0, "", true},
}

func TestMain(t *testing.T) {

	for i, test := range mockEnvVars {
		tw := timeWindow{
			startTime: *t0,
			endTime:   *t0,
		}
		fmt.Println("###############################################")
		fmt.Println("TestMain iteration ", i+1)
		tw.setTimeWindow(test.startTime, test.endTime, test.sinceTime, test.last, test.unit)
		contentSlice := parseCSV(test.files, tw)
		runFunctionality(contentSlice, test.buildsExec, test.topUsers, test.successRate, test.topExitCodes)
		if test.result != success {
			t.Fatal("Expected result not given!")
		}
	}
}
