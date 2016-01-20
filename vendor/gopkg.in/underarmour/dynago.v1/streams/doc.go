/*
Provides DynamoDB streams support for Dynago.

This provides access to the DynamoDB streams API with all the conveniences
that one would expect from using Dynago: document unmarshaling to go types,
clean API's and a way to write applications simply.

Despite what is said above, this package is still a low-level interface to the
DynamoDB streams API calls, which is a bit more clumsy to use than the higher
level interface, which is available at
http://godoc.org/github.com/crast/dynatools/streamer
*/
package streams
