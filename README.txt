Description:

Given a CSV file with the following field structure:

● Unique identifier for each build
● User ID reference the user that submitted the build request
● Time the build request was received (RFC 3339 formatted string)
● Time the build execution began (RFC 3339 formatted string)
● Time the build execution finished (RFC 3339 formatted string)
● Indicator for if the build has been deleted
● Exit code from the build process, >0 indicates failure
● Size of the resulting built image file

I have a implemented a cli tool that parses the CSV file and outputs:
For a given time window (or all entries in the file): 
- how many build were executed ("--buildsExecuted")
- which users are using the remote build service the most (top 5 users and how many builds have they executed in the time window - "--topUsers")
- the build success rate ("--successRate")
- for builds that are not succeeding what are the top exit codes

Any combination of the above options is valid (excluding the case where no option is expressed).

The time window options can be passed in number of ways:
- with the flags "--between" and "--and". Both values need to be strings in the format "mm/dd//yyyy"
ex:
go run main.go --file 1.csv -topUsers -buildsExecuted -successRate -topFailures --between 10/12/2018 --and 12/12/2018

- with the flag "--since" followed by a string that respects the "mm/dd/yyyy" format
ex:
go run main.go --file 1.csv -topUsers -buildsExecuted -successRate -topFailures --since 09/11/2018

- with the flags "--last" and "--unit" followed by an integer respectively a string (one in second/sec/seconds/secs/minute/min/minutes/mins/hour/hours/h/week/weeks/month/months)
ex:
go run main.go --file 1.csv -topUsers -buildsExecuted -successRate -topFailures --last 6 --unit months

The absence of a flag indicating the "time window" will mean that all entries in the file will be considered

Design:

The first thing in the "main" function is the parsing of the command line arguments (or flags). I have divided the flags as "time filtering" flags used to indicate the desired time window and "operational" flags, which indicate the functionality implemented on the dataset (a slice of csvLine's) resulted from the "time filtering" stage.

Based on the arguments related to the time window I've created a struct to define the time window called "timeWindow" (a method called setTimeWindow is associated with the "timeWindow" struct)

I've also created a struct called "csvLine" to hold the information associated with a individual line in the CSV file (since the CSV file is small in size it was parsed with reader.Read() - trying to implement any type of concurrency would probably slow things down for this particular case).

The "csvLines" are grouped in a slice ("contentSlice") that is created from all the lines in the CSV files who's completed build time fits in the time window ("timeWindow" struct resulted from interpreting the "time filtering" flags)

With the "contentSlice" created, we can iterate through it and determine (based on the passed in "operational" flags) the operation(s) that needs be performed. I then implement the operations and display in the console any or all of the following:

- the number of builds
- the top 5 users by number of builds
- the build success rate (based on non-zero exit codes)
- the top 5 non-zero(failure) exit codes 

Testing:

As this was not suppose to take a substantial amount of time I did not get to much done as far as testing goes. Mainly, I tried defining some end-to-end test cases resulted from mixing various types of input CSV files in the /testdata (some with empty lines or not in timely order or with a few failure exit codes), the various "time filtering" flags and a few combinations of "operational" flags. More could be added easily
To test, in the folder (/CSVParser) where main.go and main_test.go are, run:

go test

If given more time, I would:

- implement better error handling (more descriptive and helpful).
- add some more text to console so the user would be more aware of the options they have, and also explain the flags better.
- be more permissive in the way parameters are read (for ex I would also allow format mm.dd.yyyy or dd/mm/yyyy or dd.mm.yyyy to be valid)
- dd more comments to the code so others who would be looking at the code have some help in figuring out what I've done.
- implement some defer/recover functions to also make sure some test cases that require the utility to fail are enforced
- add a lot more test cases in testing.
- for larger files, implement concurrency when reading the lines in the file (replace the "lines" array from function "parseCSV" with a channel and run go routines for reading lines)
- run some test with something like "time.since" to see any bottlenecks and would try to optimize by using "benchmarking" in testing.
- maybe use "github.com/pkg/profile" to profile memory and CPU usage (if larger CSV files will need to be ingested).
