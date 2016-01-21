package dynago

type batchWriteItemRequest struct {
	RequestItems BatchWriteTableMap
}

type BatchWriteTableMap map[string][]*BatchWriteTableEntry

type BatchWriteTableEntry struct {
	DeleteRequest *batchDelete `json:",omitempty"`
	PutRequest    *batchPut    `json:",omitempty"`
}

// Set this table entry as a delete request
func (e *BatchWriteTableEntry) SetDelete(key Document) {
	e.DeleteRequest = &batchDelete{key}
}

func (e *BatchWriteTableEntry) SetPut(item Document) {
	e.PutRequest = &batchPut{item}
}

type batchDelete struct {
	Key Document
}

type batchPut struct {
	Item Document
}

type batchAction struct {
	next  *batchAction
	table string
	item  Document
}

func newBatchWrite(client *Client) *BatchWrite {
	return &BatchWrite{
		client: client,
	}
}

type BatchWrite struct {
	client  *Client
	puts    *batchAction
	deletes *batchAction
}

/*
Add some number of puts for a table.
*/
func (b BatchWrite) Put(table string, items ...Document) *BatchWrite {
	addBatchActions(&b.puts, table, items)
	return &b
}

/*
Add some number of deletes for a table.
*/
func (b BatchWrite) Delete(table string, keys ...Document) *BatchWrite {
	addBatchActions(&b.deletes, table, keys)
	return &b
}

func (b *BatchWrite) Execute() (*BatchWriteResult, error) {
	return b.client.executor.BatchWriteItem(b)
}

// Build the table map that is represented by this BatchWrite
func (b *BatchWrite) buildTableMap() (m BatchWriteTableMap) {
	m = BatchWriteTableMap{}
	ensure := func(table string) (r *BatchWriteTableEntry) {
		r = &BatchWriteTableEntry{}
		m[table] = append(m[table], r)
		return
	}

	for put := b.puts; put != nil; put = put.next {
		ensure(put.table).SetPut(put.item)
	}

	for d := b.deletes; d != nil; d = d.next {
		ensure(d.table).SetDelete(d.item)
	}
	return
}

func (e *AwsExecutor) BatchWriteItem(b *BatchWrite) (result *BatchWriteResult, err error) {
	req := batchWriteItemRequest{
		RequestItems: b.buildTableMap(),
	}

	err = e.MakeRequestUnmarshal("BatchWriteItem", req, &result)
	return
}

type BatchWriteResult struct {
	UnprocessedItems BatchWriteTableMap
	// TODO ConsumedCapacity
}

///////////////////// Batch Get

const (
	bgProjectionExpression = "ProjectionExpression"
	bgProjectionParams     = "Params"
	bgConsistentRead       = "ConsistentRead"
)

type batchGetItemRequest struct {
	RequestItems BatchGetTableMap
}

type BatchGetTableMap map[string]*BatchGetTableEntry

type BatchGetTableEntry struct {
	Keys []Document

	expressionAttributes
	ProjectionExpression string `json:",omitempty"`
	ConsistentRead       bool   `json:",omitempty"`
}

type BatchGet struct {
	client  *Client
	req     batchGetItemRequest
	gets    *batchAction
	options *batchAction
}

/*
Set which keys to get.
Can be called multiple times to queue up gets for multiple tables.
*/
func (b BatchGet) Get(table string, keys ...Document) *BatchGet {
	addBatchActions(&b.gets, table, keys)
	return &b
}

/*
Add a ProjectionExpression for a table.

Projection expression is scoped to each table, and must be called for each
table on which you want a ProjectionExpression.
*/
func (b BatchGet) ProjectionExpression(table string, expression string, params ...Params) *BatchGet {
	doc := Document{
		bgProjectionExpression: expression,
		bgProjectionParams:     params,
	}
	addBatchActions(&b.options, table, []Document{doc})
	return &b
}

/*
Set consistent read mode for a table.

Consistent read is scoped to each table, so must be called for each table in
this BatchGet for which you want consistent reads.
*/
func (b BatchGet) ConsistentRead(table string, consistent bool) *BatchGet {
	doc := Document{
		bgConsistentRead: consistent,
	}
	addBatchActions(&b.options, table, []Document{doc})
	return &b
}

func (b *BatchGet) buildTableMap() BatchGetTableMap {
	m := BatchGetTableMap{}
	ensure := func(key string) (entry *BatchGetTableEntry) {
		if entry = m[key]; entry == nil {
			entry = &BatchGetTableEntry{}
			m[key] = entry
		}
		return
	}
	for get := b.gets; get != nil; get = get.next {
		entry := ensure(get.table)
		entry.Keys = append(entry.Keys, get.item)
	}
	for option := b.options; option != nil; option = option.next {
		entry := ensure(option.table)
		for k, v := range option.item {
			switch k {
			case bgProjectionExpression:
				entry.ProjectionExpression = v.(string)
			case bgProjectionParams:
				entry.paramsHelper(v.([]Params))
			case bgConsistentRead:
				entry.ConsistentRead = v.(bool)
			}
		}
	}
	return m
}

// Execute this batch get.
func (b *BatchGet) Execute() (result *BatchGetResult, err error) {
	return b.client.executor.BatchGetItem(b)
}

func (e *AwsExecutor) BatchGetItem(b *BatchGet) (result *BatchGetResult, err error) {
	req := batchGetItemRequest{b.buildTableMap()}
	err = e.MakeRequestUnmarshal("BatchGetItem", &req, &result)
	return
}

type BatchGetResult struct {
	// Responses to the Batch Get query which were resolved.
	// Note that the order of documents is not guaranteed to be the same as
	// the order of documents requested in building this query.
	Responses map[string][]Document // table name -> list of items.

	// Unprocessed keys are keys which for some reason could not be retrieved.
	// This could be because of response size exceeding the limit or the
	// provisioned throughput being exceeded on one or more tables in this request.
	UnprocessedKeys BatchGetTableMap // Table name -> keys and settings
}

func addBatchActions(list **batchAction, table string, items []Document) {
	head := *list
	for _, item := range items {
		head = &batchAction{head, table, item}
	}
	*list = head
}
