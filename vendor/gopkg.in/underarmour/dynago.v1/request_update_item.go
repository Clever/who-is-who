package dynago

type updateItemRequest struct {
	Key       Document
	TableName string

	ConditionExpression string `json:",omitempty"`
	UpdateExpression    string `json:",omitempty"`
	expressionAttributes

	ReturnConsumedCapacity      string       `json:",omitempty"` // TODO
	ReturnItemCollectionMetrics string       `json:",omitempty"` // TODO
	ReturnValues                ReturnValues `json:",omitempty"`
}

func newUpdateItem(client *Client, table string, key Document) *UpdateItem {
	return &UpdateItem{
		client: client,
		req: updateItemRequest{
			Key:       key,
			TableName: table,
		},
	}
}

type UpdateItem struct {
	client *Client
	req    updateItemRequest
}

// Set a condition expression for conditional update.
func (u UpdateItem) ConditionExpression(expression string, params ...Params) *UpdateItem {
	u.req.paramsHelper(params)
	u.req.ConditionExpression = expression
	return &u
}

// Set an update expression to update specific fields and values.
func (u UpdateItem) UpdateExpression(expression string, params ...Params) *UpdateItem {
	u.req.paramsHelper(params)
	u.req.UpdateExpression = expression
	return &u
}

// Quick-set a single parameter
func (u UpdateItem) Param(key string, value interface{}) *UpdateItem {
	u.req.paramHelper(key, value)
	return &u
}

// Set multiple parameters at once.
func (u UpdateItem) Params(params ...Params) *UpdateItem {
	u.req.paramsHelper(params)
	return &u
}

// If set, then we will get return values of either updated or old fields (see ReturnValues const)
func (u UpdateItem) ReturnValues(returnValues ReturnValues) *UpdateItem {
	u.req.ReturnValues = returnValues
	return &u
}

// Execute this UpdateItem and return the result.
func (u *UpdateItem) Execute() (res *UpdateItemResult, err error) {
	return u.client.executor.UpdateItem(u)
}

func (e *AwsExecutor) UpdateItem(u *UpdateItem) (result *UpdateItemResult, err error) {
	if u.req.ReturnValues != ReturnNone && u.req.ReturnValues != "" {
		err = e.MakeRequestUnmarshal("UpdateItem", &u.req, &result)
	} else {
		_, err = e.makeRequest("UpdateItem", &u.req)
	}
	return
}

type UpdateItemResult struct {
	Attributes Document
}
