package dynago

/*
A Mock executor for purpose of testing.

This Executor doesn't actually run any network requests, so it can be used in
unit testing for your own application which uses Dynago.  It can be asserted
on whether a specific underlying method got called, what it was called with,
and the nature of how the query was built. You can also for most methods
control which result is returned in order to control how results make it back
to the application.

	// example, normally you'd call into a real application not inside your
	// test module and use dependency injection to specify the client.
	func application(client *dynago.Client) int {
		result, err := client.PutItem("mytable", doc).Execute()
		// do something with result maybe.
	}

	func TestApplication() {
		executor := &dynago.MockExecutor{}
		client := dynago.NewClient(executor)
		executor.PutItemResult = &dynago.PutItemResult{}

		// call into application
		application(client)

		// assert things on the executor.
		assert.Equal(true, executor.PutItemCalled)
		assert.Equal("mytable", executor.PutItemCall.Table)
		... and so on
	}

*/
type MockExecutor struct {
	Calls []MockExecutorCall // All calls made through this executor

	DeleteItemCalled bool
	DeleteItemCall   *MockExecutorCall
	DeleteItemResult *DeleteItemResult
	DeleteItemError  error

	PutItemCalled bool
	PutItemCall   *MockExecutorCall
	PutItemResult *PutItemResult
	PutItemError  error

	GetItemCalled bool
	GetItemCall   *MockExecutorCall
	GetItemResult *GetItemResult
	GetItemError  error

	BatchGetItemCalled bool
	BatchGetItemCall   *MockExecutorCall
	BatchGetItemResult *BatchGetResult

	BatchWriteItemCalled bool
	BatchWriteItemCall   *MockExecutorCall
	BatchWriteItemError  error

	QueryCalled bool              // True if query was called at least once
	QueryCall   *MockExecutorCall // Info for the last call to Query
	QueryError  error             // Specify the error from Query
	QueryResult *QueryResult      // Specify the result from Query

	ScanCalled bool
	ScanCall   *MockExecutorCall
	ScanResult *ScanResult
	ScanError  error

	UpdateItemCalled bool
	UpdateItemCall   *MockExecutorCall
	UpdateItemResult *UpdateItemResult
	UpdateItemError  error
}

// Mock executor calls
type MockExecutorCall struct {
	// used for all calls
	Method string
	Table  string

	// used for calls with expressions (most of them)
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues Document

	Key                 Document
	Item                Document
	UpdateExpression    string
	ConditionExpression string
	ReturnValues        ReturnValues
	ConsistentRead      bool

	// Query and Scan
	IndexName              string
	KeyConditionExpression string
	FilterExpression       string
	ProjectionExpression   string
	Ascending              bool
	Limit                  uint
	ExclusiveStartKey      Document
	Select                 Select
	Segment                *int
	TotalSegments          *int

	BatchWrites BatchWriteTableMap
	BatchGets   BatchGetTableMap
}

func (e *MockExecutor) BatchGetItem(batchGet *BatchGet) (*BatchGetResult, error) {
	e.BatchGetItemCalled = true
	e.addCall(&e.BatchGetItemCall, MockExecutorCall{
		Method:    "BatchGetItem",
		BatchGets: batchGet.buildTableMap(),
	})
	return e.BatchGetItemResult, nil
}

func (e *MockExecutor) BatchWriteItem(batchWrite *BatchWrite) (*BatchWriteResult, error) {
	e.BatchWriteItemCalled = true
	e.addCall(&e.BatchWriteItemCall, MockExecutorCall{
		Method:      "BatchWriteItem",
		BatchWrites: batchWrite.buildTableMap(),
	})

	return &BatchWriteResult{}, e.BatchWriteItemError
}

func (e *MockExecutor) DeleteItem(deleteItem *DeleteItem) (*DeleteItemResult, error) {
	e.DeleteItemCalled = true
	e.addCall(&e.DeleteItemCall, MockExecutorCall{
		Method:                    "DeleteItem",
		Table:                     deleteItem.req.TableName,
		Key:                       deleteItem.req.Key,
		ConditionExpression:       deleteItem.req.ConditionExpression,
		ExpressionAttributeNames:  deleteItem.req.ExpressionAttributeNames,
		ExpressionAttributeValues: deleteItem.req.ExpressionAttributeValues,
		ReturnValues:              deleteItem.req.ReturnValues,
	})
	return e.DeleteItemResult, e.DeleteItemError
}

func (e *MockExecutor) GetItem(getItem *GetItem) (*GetItemResult, error) {
	e.GetItemCalled = true
	call := MockExecutorCall{
		Method:                    "GetItem",
		Table:                     getItem.req.TableName,
		Key:                       getItem.req.Key,
		ConsistentRead:            getItem.req.ConsistentRead,
		ExpressionAttributeValues: getItem.req.ExpressionAttributeValues,
		ExpressionAttributeNames:  getItem.req.ExpressionAttributeNames,
		ProjectionExpression:      getItem.req.ProjectionExpression,
	}
	e.GetItemCall = &call
	e.Calls = append(e.Calls, call)
	return e.GetItemResult, e.GetItemError
}

func (e *MockExecutor) PutItem(putItem *PutItem) (*PutItemResult, error) {
	e.PutItemCalled = true
	call := MockExecutorCall{
		Method:                    "PutItem",
		Table:                     putItem.req.TableName,
		Item:                      putItem.req.Item,
		ReturnValues:              putItem.req.ReturnValues,
		ConditionExpression:       putItem.req.ConditionExpression,
		ExpressionAttributeNames:  putItem.req.ExpressionAttributeNames,
		ExpressionAttributeValues: putItem.req.ExpressionAttributeValues,
	}
	e.PutItemCall = &call
	e.Calls = append(e.Calls, call)
	return e.PutItemResult, e.PutItemError
}

func callFromQueryReq(req queryRequest) MockExecutorCall {
	ascending, consistent := true, false
	if req.ScanIndexForward != nil {
		ascending = *req.ScanIndexForward
	}
	if req.ConsistentRead != nil {
		consistent = *req.ConsistentRead
	}

	return MockExecutorCall{
		Method:                    "Query",
		Table:                     req.TableName,
		IndexName:                 req.IndexName,
		KeyConditionExpression:    req.KeyConditionExpression,
		FilterExpression:          req.FilterExpression,
		ExpressionAttributeNames:  req.ExpressionAttributeNames,
		ExpressionAttributeValues: req.ExpressionAttributeValues,
		ProjectionExpression:      req.ProjectionExpression,
		Select:                    req.Select,
		Ascending:                 ascending,
		ConsistentRead:            consistent,
		Limit:                     req.Limit,
		ExclusiveStartKey:         req.ExclusiveStartKey,
	}
}

func (e *MockExecutor) Query(query *Query) (*QueryResult, error) {
	e.QueryCalled = true
	e.addCall(&e.QueryCall, callFromQueryReq(query.req))

	result := e.QueryResult
	if result != nil {
		result.query = query
		if result.Count == 0 {
			result.Count = len(result.Items)
		}
		if result.ScannedCount == 0 {
			result.ScannedCount = result.Count
		}
	}

	return result, e.QueryError
}

func (e *MockExecutor) Scan(scan *Scan) (*ScanResult, error) {
	e.ScanCalled = true
	call := callFromQueryReq(scan.req.queryRequest)
	call.Method = "Scan"
	call.Segment = scan.req.Segment
	call.TotalSegments = scan.req.TotalSegments
	e.addCall(&e.ScanCall, call)
	return e.ScanResult, e.ScanError
}

func (e *MockExecutor) UpdateItem(update *UpdateItem) (*UpdateItemResult, error) {
	e.UpdateItemCalled = true
	e.addCall(&e.UpdateItemCall, MockExecutorCall{
		Method:                    "UpdateItem",
		Table:                     update.req.TableName,
		UpdateExpression:          update.req.UpdateExpression,
		ConditionExpression:       update.req.ConditionExpression,
		ExpressionAttributeNames:  update.req.ExpressionAttributeNames,
		ExpressionAttributeValues: update.req.ExpressionAttributeValues,
	})
	return e.UpdateItemResult, e.UpdateItemError
}

// Currently we don't implement mocking for SchemaExecutor. Returns nil.
func (e *MockExecutor) SchemaExecutor() SchemaExecutor {
	return nil
}

// Reduce boilerplate on adding a call
func (e *MockExecutor) addCall(target **MockExecutorCall, call MockExecutorCall) {
	e.Calls = append(e.Calls, call)
	if target != nil {
		*target = &call
	}
}

// Convenience method to get delete keys in this batch write map for a specific table
func (m BatchWriteTableMap) GetDeleteKeys(table string) (deletes []Document) {
	for _, entry := range m[table] {
		if entry.DeleteRequest != nil {
			deletes = append(deletes, entry.DeleteRequest.Key)
		}
	}
	return
}

// Convenience method to get all put documents in this batch write map for a specific table
func (m BatchWriteTableMap) GetPuts(table string) (puts []Document) {
	for _, entry := range m[table] {
		if entry.PutRequest != nil {
			puts = append(puts, entry.PutRequest.Item)
		}
	}
	return
}
