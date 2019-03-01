package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"
)

// the time.Time zero value
var t0 = new(time.Time)
var success = false

// struct to hold each line
type csvLine struct {
	BuildID            string
	User               string
	BuildReqTime       time.Time
	BuildExecStartTime time.Time
	BuildExecEndTime   time.Time
	BuildDeleted       bool
	BuildExitCode      string
	BuildSize          string
}

// struct to hold time window
type timeWindow struct {
	startTime time.Time
	endTime   time.Time
}

func (t *timeWindow) setTimeWindow(startT string, endT string, sinceT string, last int64, unit string) {
	fmt.Println("-------------------------------------")
	oTime := convertTimeFormat(startT)
	nTime := convertTimeFormat(endT)
	tSince := convertTimeFormat(sinceT)
	if (oTime == *t0 && nTime == *t0) && tSince == *t0 && (last == 0 && unit == "") {
		fmt.Println("No indication of time window was given, considering all samples in the file")
		t.startTime = *t0
		t.endTime = *t0
	} else if (oTime != *t0 && nTime != *t0) && tSince == *t0 && (last == 0 && unit == "") {
		fmt.Printf("Chosen a time window between %v and %v .\n", oTime, nTime)
		t.startTime = oTime
		t.endTime = nTime
	} else if (oTime == *t0 && nTime == *t0) && tSince != *t0 && (last == 0 && unit == "") {
		now := time.Now().Format(time.RFC3339)
		fTime, err := time.Parse(time.RFC3339, now)
		if err != nil {
			panic(fmt.Sprintf("Could not parse \"current time\" %s", err))
		}
		fmt.Printf("Chosen a time window between %v and %v (now).\n", tSince, now)
		t.startTime = tSince
		t.endTime = fTime
	} else if (oTime == *t0 && nTime == *t0) && tSince == *t0 && (last != 0 && unit != "") {
		now := time.Now()
		nrUnits := last
		timeUnit := unit
		var startTime int64
		if timeUnit == "seconds" || timeUnit == "second" || timeUnit == "secs" || timeUnit == "sec" || timeUnit == "s" {
			startTime = now.Unix() - nrUnits
		} else if timeUnit == "minutes" || timeUnit == "minute" || timeUnit == "mins" || timeUnit == "min" {
			startTime = now.Unix() - nrUnits*60
		} else if timeUnit == "hours" || timeUnit == "hour" || timeUnit == "h" {
			startTime = now.Unix() - nrUnits*60*60
		} else if timeUnit == "days" || timeUnit == "day" || timeUnit == "d" {
			startTime = now.Unix() - nrUnits*60*60*24
		} else if timeUnit == "weeks" || timeUnit == "week" || timeUnit == "w" {
			startTime = now.Unix() - nrUnits*60*60*24*7
		} else if timeUnit == "months" || timeUnit == "month" {
			startTime = now.Unix() - nrUnits*60*60*24*30
		} else {
			panic(fmt.Sprintf("Unknown time unit!"))
		}
		sT := time.Unix(startTime, 0).Format(time.RFC3339)
		fmt.Printf("Chosen a time window that goes back %d %s (since %s).\n", nrUnits, timeUnit, sT)
		fTime, err := time.Parse(time.RFC3339, sT)
		if err != nil {
			panic(fmt.Sprintf("Could not parse \"from time\": %s", err))
		}
		current, err := time.Parse(time.RFC3339, now.Format(time.RFC3339))
		if err != nil {
			panic(fmt.Sprintf("Could not parse \"current time\": %s", err))
		}
		t.startTime = fTime
		t.endTime = current
	} else {
		panic(fmt.Sprintf("Incompatible time options chosen!"))
	}
}

// func to conv strings to time.Time (RFC3339)
func convertTimeFormat(s string) time.Time {
	if s != "" {
		initT, err := time.Parse("01/02/2006", s)
		if err != nil {
			panic(err)
		}
		sT := initT.Format(time.RFC3339)
		t, err := time.Parse(time.RFC3339, sT)
		if err != nil {
			panic(err)
		}
		return t
	}
	t := new(time.Time)
	return *t
}

// func to revert sort and get top 5 most frequent events
func sortTop5(m map[string]int, s string) {
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	i := 1
	for _, kv := range ss {
		if i == 6 {
			break
		}
		fmt.Printf("%d.\"%s\", %s: %d\n", i, kv.Key, s, kv.Value)
		i++
	}
}

func readCSVLine(pC *[]csvLine, reader *csv.Reader, tw timeWindow) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic occured: ", r)
			fmt.Println("Recovered")
		}
	}()
	content := *pC
	line, err := reader.Read()
	if err == io.EOF {
		return true
	} else if err != nil {
		panic(fmt.Sprintf("There was a problem reading the line! :. %s", err))
	}
	execEndTime, err := time.Parse(time.RFC3339, line[4])
	if err != nil {
		panic(fmt.Sprintf("Could not parse build time!: %s", err))
	}
	bAfterLowerBound := execEndTime.After(tw.startTime)
	bBeforeHigherBound := execEndTime.Before(tw.endTime)
	if (bAfterLowerBound && bBeforeHigherBound) || (tw.startTime == *t0 && tw.endTime == *t0) {
		reqTime, err := time.Parse(time.RFC3339, line[2])
		if err != nil {
			panic(fmt.Sprintf("Could not parse \"request time\" (the 3rd field in line): %s", err))
		}
		execStartTime, err := time.Parse(time.RFC3339, line[3])
		if err != nil {
			panic(fmt.Sprintf("Could not parse \"execution time\" (the 4th field in line): %s", err))
		}
		bDeleted, err := strconv.ParseBool(line[5])
		if err != nil {
			panic(fmt.Sprintf("Could not parse \"build deleted\" (the 6th field in line): %s", err))
		}
		content = append(content, csvLine{
			BuildID:            line[0],
			User:               line[1],
			BuildReqTime:       reqTime,
			BuildExecStartTime: execStartTime,
			BuildExecEndTime:   execEndTime,
			BuildDeleted:       bDeleted,
			BuildExitCode:      line[6],
			BuildSize:          line[7],
		})
	}
	*pC = content
	return false
}

// create a slice to hold the structs that contain the data for each line
func parseCSV(s string, tw timeWindow) []csvLine {
	// open CSV file
	file, err := os.Open(s)
	if err != nil {
		panic(fmt.Sprintf("There was a problem opening the file! :. %s", err))
	}
	// make sure we eventually close the CSV file
	defer func() {
		if err = file.Close(); err != nil {
			panic(fmt.Sprintf("There was a problem closing the file! :. %s", err))
		}
	}()
	// create a reader for the file
	reader := csv.NewReader(file)
	var content []csvLine

	// if the read line is in the specified time window put it into the content slice
	for {
		EOF := readCSVLine(&content, reader, tw)
		if EOF {
			break
		}
	}
	return content
}

func runFunctionality(content []csvLine, buildsExec bool, topUsers bool, successRate bool, topExitCodes bool) {
	if len(content) == 0 {
		panic(fmt.Sprintf("There are no etries in the CSV file matching the chosen time window."))
	}
	if buildsExec {
		fmt.Println("-------------------------------------")
		fmt.Println("Chosen to see how many builds were executed in the relevant time window. ")
		fmt.Println("The number of executed builds is: ", len(content))
		fmt.Println("-------------------------------------")
	}
	if topUsers {
		fmt.Println("Chosen to see who are the top 5 users in the relevant time window. ")
		fmt.Println("The top 5 Users are:")
		duplicateFrequency := make(map[string]int)
		for _, item := range content {
			_, exist := duplicateFrequency[item.User]
			if exist {
				// if already in the map increase counter by 1
				duplicateFrequency[item.User]++
			} else {
				// else start from 1
				duplicateFrequency[item.User] = 1
			}
		}
		sortTop5(duplicateFrequency, "Builds")
		fmt.Println("-------------------------------------")
	}
	if successRate {
		fmt.Println("Chosen to see what the build suceess rate was in the relevant time window. ")
		var successes int
		for _, item := range content {
			if item.BuildExitCode == "0" {
				successes = successes + 1
			}
		}
		sRate := successes * 100 / len(content)
		fmt.Printf("The build success rate is: %d %% \n", sRate)
		fmt.Println("-------------------------------------")
	}
	if topExitCodes {
		fmt.Println("Chosen to see what are the top 5 exit codes for failed builds in the relevant time window. ")
		fmt.Println("The top 5 failure exit codes are:")
		duplicateFrequency := make(map[string]int)
		for _, item := range content {
			if item.BuildExitCode != "0" {
				_, exist := duplicateFrequency[item.BuildExitCode]
				if exist {
					// if already in the map increase counter by 1
					duplicateFrequency[item.BuildExitCode]++
				} else {
					// else start from 1
					duplicateFrequency[item.BuildExitCode] = 1
				}
			}
		}
		if len(duplicateFrequency) == 0 {
			fmt.Println("No failed builds.")
		} else {
			sortTop5(duplicateFrequency, "Occurrences")
		}
		fmt.Println("-------------------------------------")
		success = true
	}
}

func main() {
	//get and deal with args from cli
	fileArg := flag.String("file", "", "Path of CSV file to be parsed. (Required)")
	buildsExec := flag.Bool("buildsExecuted", false, "Reports the number of builds executed in a specfic time window")
	topUsers := flag.Bool("topUsers", false, "List the top 5 users by number of builds and how many builds they have executed in a specific time window ")
	successRate := flag.Bool("successRate", false, "Percentage of successful builds in a specific time window")
	topExitCodes := flag.Bool("topFailures", false, "List the top 5 exit codes for unsucessful builds in a specific time window ")
	oldTime := flag.String("between", "", "Specify the older time window margin (requires \"-and\")")
	newTime := flag.String("and", "", "Specify the newer time window margin  (requires \"-between\")")
	since := flag.String("since", "", "Specify a point in time since when the reporting should start")
	last := flag.Int64("last", 0, "Specify the number of units of time (requires \"--unit\") since you would like the reporting to start")
	unit := flag.String("unit", "", "Specify the unit of time (requires \"--last\") you would like the reporting to start (choos from: minutes, hours, days, weeks, months)")
	flag.Parse()

	if *fileArg == "" {
		panic(fmt.Sprintf("Please use the \"-file\" option to specify a CSV file!"))
	}

	if !*buildsExec && !*topUsers && !*successRate && !*topExitCodes {
		panic(fmt.Sprintf("No action was chosen! Please indicate an action that you would like performed!"))
	}

	var tw timeWindow
	// set the time window according to cli flags
	tw.setTimeWindow(*oldTime, *newTime, *since, *last, *unit)

	// create a slice for each CSV line
	contentSlice := parseCSV(*fileArg, tw)

	// run indicated operations and display results
	runFunctionality(contentSlice, *buildsExec, *topUsers, *successRate, *topExitCodes)
}
