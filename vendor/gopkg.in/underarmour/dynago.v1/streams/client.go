package streams

const targetPrefix = "DynamoDBStreams_20120810." // This is the Dynamo API version we support

type Config struct {
	// Must implement MakeRequestUnmarshal. The easiest thing to use in this
	// case is an instance of dynago.AwsExecutor or the like.
	Requester MakeRequester
}

func NewClient(config *Config) *Client {
	return &Client{config.Requester}
}

/*
Client is the low-level interface to the streams API.
Here, all the individual Streams API requests are available.
*/
type Client struct {
	caller MakeRequester
}

func (s *Client) DescribeStream(request *DescribeStreamRequest) (dest *DescribeStreamResponse, err error) {
	err = s.caller.MakeRequestUnmarshal(targetPrefix+"DescribeStream", request, &dest)
	return
}

func (s *Client) GetShardIterator(req *GetIteratorRequest) (dest *GetIteratorResult, err error) {
	err = s.caller.MakeRequestUnmarshal(targetPrefix+"GetShardIterator", req, &dest)
	return
}

func (s *Client) GetRecords(req *GetRecordsRequest) (result *GetRecordsResponse, err error) {
	err = s.caller.MakeRequestUnmarshal(targetPrefix+"GetRecords", req, &result)
	return
}

type MakeRequester interface {
	MakeRequestUnmarshal(method string, document interface{}, dest interface{}) (err error)
}
