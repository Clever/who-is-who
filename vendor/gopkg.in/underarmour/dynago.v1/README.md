Dynago
======

[![Build Status](https://travis-ci.org/underarmour/dynago.svg?branch=master)](https://travis-ci.org/underarmour/dynago) [![GoDoc](https://godoc.org/gopkg.in/underarmour/dynago.v1?status.svg)](https://godoc.org/gopkg.in/underarmour/dynago.v1)

Dynago is a DynamoDB client API for Go.

This attempts to be a really simple, principle of least-surprise client for the DynamoDB API.

Key design tenets of Dynago:

 * Most actions are done via chaining to build filters and conditions
 * objects are completely safe for passing between goroutines (even queries and the like)
 * To make understanding easier via docs, we use amazon's naming wherever possible.

Installation
------------
Install using `go get`:

    go get gopkg.in/underarmour/dynago.v1

Docs are at http://godoc.org/gopkg.in/underarmour/dynago.v1

Example
-------

Run a query:

```go
client := dynago.NewClient(endpoint, accessKey, secretKey)

query := client.Query(table).
	KeyConditionExpression("UserId = :uid", dynago.Param{":uid", 42}).
	FilterExpression("NumViews > :views").
	Param(":views", 50).
	Desc()

result, err := query.Execute()
if err != nil {
	// do something
}
for _, row := range result.Items {
	fmt.Printf("Name: %s, Views: %d", row["Name"], row["NumViews"])
}
```

Type Marshaling
---------------

Dynago lets you use go types instead of having to understand a whole lot about dynamo's internal type system.

Example:

```go
doc := dynago.Document{
	"name": "Bob",
	"age": 45,
	"height": 2.1,
	"address": dynago.Document{
		"city": "Boston",
	},
	"tags": dynago.StringSet{"male", "middle_aged"},
}
client.PutItem("person", doc).Execute()
```

 * Strings use golang `string`
 * Numbers can be input as `int` (`int64`, `uint64`, etc) or `float64` but always are returned as [`dynago.Number`][dynagoNumber] to not lose precision.
 * Maps can be either `map[string]interface{}` or [`dynago.Document`][dynagoDocument]
 * Opaque binary data can be put in `[]byte`
 * String sets, number sets, binary sets are supported using [`dynago.StringSet`][dynagoStringSet] `dynago.NumberSet` `dynago.BinarySet`
 * Lists are supported using [`dynago.List`][dynagoList]
 * `time.Time` is only accepted if it's a UTC time, and is marshaled to a dynamo string in iso8601 compact format. It comes back as a string, an can be got back using `GetTime()` on `Document`.

[dynagoDocument]: http://godoc.org/gopkg.in/underarmour/dynago.v1#Document
[dynagoList]: http://godoc.org/gopkg.in/underarmour/dynago.v1#List
[dynagoNumber]: http://godoc.org/gopkg.in/underarmour/dynago.v1#Number
[dynagoStringSet]: http://godoc.org/gopkg.in/underarmour/dynago.v1#StringSet

Debugging
---------

Dynago can dump request or response information for you for use in debugging.
Simply set [`dynago.Debug`][dynagoDebug] with the necessary flags:

```go
dynago.Debug = dynago.DebugRequests | dynago.DebugResponses
```

If you would like to change how the debugging is printed, please set [`dynago.DebugFunc`][dynagoDebugFunc] to your preference.

[dynagoDebug]: http://godoc.org/gopkg.in/underarmour/dynago.v1#Debug
[dynagoDebugFunc]: http://godoc.org/gopkg.in/underarmour/dynago.v1#DebugFunc

Version Compatibility
---------------------

Dynago follows [Semantic Versioning](http://semver.org/) via the gopkg.in interface, and within the v1 chain, we will not break the existing API or behaviour of existing code using Dynago. We will add new methods and features, but it again should not break code.

Additional resources
--------------------
 * [DynamoDB's own API reference][apireference] explains the operations that DynamoDB supports, and as such will provide more information on how specific parameters and values within dynago actually work.
 * http://godoc.org/github.com/crast/dynatools is a collection of packages with "edge" functionality for Dynago, which includes additional libraries to add on, and some functionality fixes which may be considered for merging into dynago core in the future. It includes bits such as pluggable authentication, [support for DynamoDB streams](http://godoc.org/github.com/crast/dynatools/streamer#Streamer), [safe update expressions](http://godoc.org/github.com/crast/dynatools/safeupdate) and more.

[apireference]: http://docs.aws.amazon.com/amazondynamodb/latest/APIReference/Welcome.html

The past, and the future
------------------------

Dynago currently implements all of its support for the underlying DynamoDB API encoding, AWS signing/authentication, etc. This happened in part because the existing libraries out there at the time of writing used deprecated API's and complicated methods, and it was actually cleaner at the time to support the API by fresh implementation.

[AWS-SDK-Go](https://github.com/aws/aws-sdk-go) exists as of June 2015 and has a very up to date API, but it also comes with the pain of using bare structs which minimally wrap protocol-level details of DynamoDB, which makes it a pain to use for writing applications (dealing with DynamoDB's internal type system is boilerplatey). Once Amazon has brought it out of developer preview, the plan is to have Dynago use it as the underlying protocol and signature implementation, but keep providing Dynago's clean and simple API for building queries and marshaling datatypes in DynamoDB.
