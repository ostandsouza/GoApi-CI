# GoApi-CI

GoApi-CI is an golang based api which is used to communicate with mongoDB(GridFS) backend. This is the backend for ReactHub project.
You can checkout ReactHub for more details:
https://github.com/ostandsouza/ReactHub

## Getting Started:

1) Download all go depedancies

2) Pull go Api project and run locally with help of go command.
eg: go run <file_name>

## Samples
These are few sample containing mongoDB sample data

### Collections
All suites should have its own collection in mongoDB eg:DockerExpress,dcokerProtrade. 
To store all bug id for suite in key value format we use collection: "IssueMap".
To store results files we use gridFS for this eg: "file.files", which in turn storage as chunks of data eg:"file.chunks".

> show collections

DockerExpress
IssueMap
MongoRest
dcokerProtrade
dockerExpress
file.chunks
file.files

### Testdata sample
This is how test data api looks like in database, after testdata api is called i.e once test script has executed successfully.

> db.dockerExpress.find()

{ "_id" : ObjectId("5d3f04825307204b2eb44a07"), "time" : "08:30:00", "suitename" : "DockerExpress", "device" : "opera", "notes" : "galen+selenium", "FailedTestCases" : "11", "PassedTestCases" : "1", "RunType" : "Smoke", "SkippedTestCases" : "0", "TotalTestCases" : "12", "functionalReport" : "5d3f0104405c424636b19713", "uiReport" : "5d3f02dd671e10db00ed29ae", "version" : "3.15" }

### TestResult sample
This is how test results api looks like in database, after test results api is called i.e once test script has executed successfully.

> db.file.files.find()

{ "_id" : ObjectId("5d30b1f3f728d55e3b84a1f9"), "length" : NumberLong(19), "chunkSize" : 261120, "uploadDate" : ISODate("2019-07-18T17:52:53.045Z"), "filename" : "Log_TS00099_2.txt", "metadata" : { "content-type" : "application/json" } }

### IssueMap sample
This is an optional collection, containing bug id for all suites which mapped in key value format.

{ "_id" : ObjectId("5d2aff366d0a8e0c94559f1c"), "dcokerProtrade" : [ "NEST-1234", "NEST-2345" ], "dockerExpress" : [ "NEST-4567" ], "MongoRest" : [ "NEST-7890", "NEST-5678", "NEST-3456" ] }

## Authors
Ostan Dsouza - Initial work

