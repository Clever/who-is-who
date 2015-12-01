package dynago

type deleteItemRequest struct {
	TableName string
	Key       Document

	ConditionExpression string `json:",omitempty"`
	expressionAttributes

	// TODO ReturnConsumedCapacity string
	// TODO ReturnItemCollectionMetrics
	ReturnValues ReturnValues `json:",omitempty"`
}

func newDeleteItem(client *Client, table string, key Document) *DeleteItem {
	return &DeleteItem{
		executor: client.executor,
		req: deleteItemRequest{
			TableName: table,
			Key:       key,
		},
	}
}

type DeleteItem struct {
	executor Executor
	req      deleteItemRequest
}

// Set a ConditionExpression to do a conditional DeleteItem.
func (d DeleteItem) ConditionExpression(expression string, params ...Params) *DeleteItem {
	d.req.ConditionExpression = expression
	d.req.paramsHelper(params)
	return &d
}

// Set ReturnValues. For DeleteItem, it can only be ReturnAllOld
func (d DeleteItem) ReturnValues(returnValues ReturnValues) *DeleteItem {
	d.req.ReturnValues = returnValues
	return &d
}

/*
Actually Execute this putitem.

DeleteItemResult will be nil unless ReturnValues is set.
*/
func (d *DeleteItem) Execute() (res *DeleteItemResult, err error) {
	return d.executor.DeleteItem(d)
}

func (e *AwsExecutor) DeleteItem(d *DeleteItem) (res *DeleteItemResult, err error) {
	if d.req.ReturnValues != ReturnNone && d.req.ReturnValues != "" {
		err = e.MakeRequestUnmarshal("DeleteItem", &d.req, &res)
	} else {
		_, err = e.makeRequest("DeleteItem", &d.req)
	}
	return
}

type DeleteItemResult struct {
	Attributes Document
}
