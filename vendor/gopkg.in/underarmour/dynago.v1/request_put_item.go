package dynago

type putItemRequest struct {
	TableName string
	Item      Document

	ConditionExpression string `json:",omitempty"`
	expressionAttributes

	// TODO ReturnConsumedCapacity string
	// TODO ReturnItemCollectionMetrics
	ReturnValues ReturnValues `json:",omitempty"`
}

func newPutItem(client *Client, table string, item Document) *PutItem {
	return &PutItem{
		client: client,
		req: putItemRequest{
			TableName: table,
			Item:      item,
		},
	}
}

type PutItem struct {
	client *Client
	req    putItemRequest
}

// Set a ConditionExpression to do a conditional PutItem.
func (p PutItem) ConditionExpression(expression string, params ...Params) *PutItem {
	p.req.ConditionExpression = expression
	p.req.paramsHelper(params)
	return &p
}

// Set parameter for ConditionExpression
func (p PutItem) Param(key string, value interface{}) *PutItem {
	p.req.paramHelper(key, value)
	return &p
}

func (p PutItem) Params(params ...Params) *PutItem {
	p.req.paramsHelper(params)
	return &p
}

// Set ReturnValues.
func (p PutItem) ReturnValues(returnValues ReturnValues) *PutItem {
	p.req.ReturnValues = returnValues
	return &p
}

/*
Actually Execute this putitem.

PutItemResult will be nil unless ReturnValues or ReturnConsumedCapacity is set.
*/
func (p *PutItem) Execute() (res *PutItemResult, err error) {
	return p.client.executor.PutItem(p)
}

func (e *AwsExecutor) PutItem(p *PutItem) (res *PutItemResult, err error) {
	if p.req.ReturnValues != ReturnNone && p.req.ReturnValues != "" {
		err = e.MakeRequestUnmarshal("PutItem", &p.req, &res)
	} else {
		_, err = e.makeRequest("PutItem", &p.req)
	}
	return
}

type PutItemResult struct {
	Attributes Document
}
