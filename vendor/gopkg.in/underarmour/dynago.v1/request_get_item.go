package dynago

type getItemRequest struct {
	TableName string

	Key Document

	ProjectionExpression string `json:",omitempty"`
	expressionAttributes

	// TODO ReturnConsumedCapacity string
	ConsistentRead bool `json:",omitempty"`
}

func newGetItem(client *Client, table string, key Document) *GetItem {
	return &GetItem{
		client: client,
		req: getItemRequest{
			TableName: table,
			Key:       key,
		},
	}
}

type GetItem struct {
	client *Client
	req    getItemRequest
}

// Set the ProjectionExpression for this GetItem (which attributes to get)
func (p GetItem) ProjectionExpression(expression string) *GetItem {
	p.req.ProjectionExpression = expression
	return &p
}

// Shortcut to set an ExpressionAttributeValue for used in expression query
func (p GetItem) Param(key string, value interface{}) *GetItem {
	p.req.paramHelper(key, value)
	return &p
}

func (p GetItem) Params(params ...Params) *GetItem {
	p.req.paramsHelper(params)
	return &p
}

// Set up this get to be a strongly consistent read.
func (p GetItem) ConsistentRead(strong bool) *GetItem {
	p.req.ConsistentRead = strong
	return &p
}

// Execute the get item.
func (p *GetItem) Execute() (result *GetItemResult, err error) {
	return p.client.executor.GetItem(p)
}

func (e *AwsExecutor) GetItem(g *GetItem) (result *GetItemResult, err error) {
	err = e.MakeRequestUnmarshal("GetItem", &g.req, &result)
	return
}

// The result from executing a GetItem.
type GetItemResult struct {
	Item             Document
	ConsumedCapacity interface{} // TODO
}
