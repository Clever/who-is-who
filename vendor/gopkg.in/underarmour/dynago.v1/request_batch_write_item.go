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
	b.addActions(&b.puts, table, items)
	return &b
}

/*
Add some number of deletes for a table.
*/
func (b BatchWrite) Delete(table string, keys ...Document) *BatchWrite {
	b.addActions(&b.deletes, table, keys)
	return &b
}

func (b *BatchWrite) addActions(list **batchAction, table string, items []Document) {
	head := *list
	for _, item := range items {
		head = &batchAction{head, table, item}
	}
	*list = head
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
