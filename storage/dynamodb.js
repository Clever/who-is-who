const { DynamoDBClient, DescribeTableCommand, CreateTableCommand } = require("@aws-sdk/client-dynamodb");
const { DynamoDBDocumentClient, GetCommand, ScanCommand, QueryCommand, PutCommand, DeleteCommand } = require("@aws-sdk/lib-dynamodb");
const async = require("async");

const _ = require("lodash");
const kv = require("kayvee");

const log = new kv.logger("who-is-who");

const objTable = {
  TableName: "whoswho-objects",
  AttributeDefinitions: [{ AttributeName: "_whoid", AttributeType: "S" }],
  KeySchema: [{ AttributeName: "_whoid", KeyType: "HASH" }],
  ProvisionedThroughput: { ReadCapacityUnits: 5, WriteCapacityUnits: 5 },
};
const pathTable = {
  TableName: "whoswho-paths",
  AttributeDefinitions: [
    { AttributeName: "path", AttributeType: "S" },
    { AttributeName: "val_whoid", AttributeType: "S" },
  ],
  KeySchema: [
    { AttributeName: "path", KeyType: "HASH" },
    { AttributeName: "val_whoid", KeyType: "RANGE" },
  ],
  ProvisionedThroughput: { ReadCapacityUnits: 5, WriteCapacityUnits: 5 },
};
const histTable = {
  TableName: "whoswho-history",
  AttributeDefinitions: [
    { AttributeName: "_whoid", AttributeType: "S" },
    { AttributeName: "path_time", AttributeType: "S" },
  ],
  KeySchema: [
    { AttributeName: "_whoid", KeyType: "HASH" },
    { AttributeName: "path_time", KeyType: "RANGE" },
  ],
  ProvisionedThroughput: { ReadCapacityUnits: 5, WriteCapacityUnits: 5 },
};

function checkSchema(expected, actual) {
  let minactual = {
    TableName: actual.TableName,
    AttributeDefinitions: _.sortBy(actual.AttributeDefinitions, "AttributeName"),
    KeySchema: _.sortBy(actual.KeySchema, "AttributeName"),
    ProvisionedThroughput: {
      ReadCapacityUnits: actual.ProvisionedThroughput.ReadCapacityUnits,
      WriteCapacityUnits: actual.ProvisionedThroughput.WriteCapacityUnits,
    },
  };

  if (_.isEqual(expected, minactual)) {
    return null;
  } else {
    return new Error(
      "Mismatched table schema for " +
        expected.TableName +
        "\n" +
        "expected:\n" +
        JSON.stringify(expected, null, 4) +
        "\n" +
        "actual:\n" +
        JSON.stringify(minactual, null, 4),
    );
  }
}

function createTableIfNeeded(dynamodb, table, cb) {
  dynamodb.send(new DescribeTableCommand({ TableName: table.TableName }))
    .then(data => cb(checkSchema(table, data.Table)))
    .catch(err => {
      if (err.name === "ResourceNotFoundException") {
        log.warn("creating table " + table.TableName);
        dynamodb.send(new CreateTableCommand(table))
          .then(() => cb(null))
          .catch(cb);
      } else {
        cb(err);
      }
    });
}

function createTablesIfNeeded(endpoint, region, accessId, secretKey, cb) {
  const dynamodb = new DynamoDBClient({
    endpoint: endpoint,
    region: region,
    credentials: { accessKeyId: accessId, secretAccessKey: secretKey },
  });

  async.parallel(
    [
      cb => createTableIfNeeded(dynamodb, objTable, cb),
      cb => createTableIfNeeded(dynamodb, pathTable, cb),
      cb => createTableIfNeeded(dynamodb, histTable, cb),
    ],
    cb
  );
}

function createWorkingExport(endpoint, region, accessId, secretKey) {
  const rawClient = new DynamoDBClient({
    endpoint: endpoint,
    region: region,
    credentials: { accessKeyId: accessId, secretAccessKey: secretKey },
  });
  const client = DynamoDBDocumentClient.from(rawClient);

  const toObj = (obj, cb) => {
    client.send(new GetCommand({ TableName: objTable.TableName, Key: { _whoid: obj._whoid } }))
      .then(data => cb(null, data.Item))
      .catch(err => cb(err));
  };

  return {
    all(cb) {
      client.send(new ScanCommand({ TableName: objTable.TableName }))
        .then(data => cb(null, data.Items))
        .catch(err => cb(err));
    },
    has(key, cb) {
      const params = {
        TableName: pathTable.TableName,
        KeyConditionExpression: "#path = :key",
        ExpressionAttributeNames: { "#path": "path" },
        ExpressionAttributeValues: { ":key": key },
      };

      client.send(new QueryCommand(params))
        .then(data => async.map(data.Items, toObj, cb))
        .catch(err => cb(err));
    },
    by(key, val, cb) {
      const params = {
        TableName: pathTable.TableName,
        KeyConditionExpression: "#path = :key and begins_with(val_whoid, :val)",
        ExpressionAttributeNames: { "#path": "path" },
        ExpressionAttributeValues: { ":key": key, ":val": val + "\u0000" },
      };

      client.send(new QueryCommand(params))
        .then(data => async.map(data.Items, toObj, cb))
        .catch(err => cb(err));
    },
    history(whoid, path, cb) {
      const params = {
        TableName: histTable.TableName,
        KeyConditionExpression: "#whoid = :whoid",
        ExpressionAttributeNames: { "#whoid": "_whoid" },
        ExpressionAttributeValues: { ":whoid": whoid },
      };

      if (path) {
        params.KeyConditionExpression += " and begins_with(path_time, :path)";
        params.ExpressionAttributeValues[":path"] = path + ".";
      }

      client.send(new QueryCommand(params))
        .then(data => {
          const results = data.Items.map(d => _.omit(d, ["path_time"]));
          cb(null, results);
        })
        .catch(err => cb(err));
    },
    put(cur, diffs, cb) {
      const callbacks = [];
      const whoid = cur._whoid;
      const now = new Date();

      callbacks.push(cb => {
        client.send(new PutCommand({ TableName: objTable.TableName, Item: cur }))
          .then(() => cb(null))
          .catch(err => cb(err));
      });

      Object.keys(diffs).forEach(path => {
        const diff = diffs[path];

        callbacks.push(cb => {
          if (diff.created) return cb(null);
          const prefix = _.isEqual(diff.prev, {}) ? "" : diff.prev + "\u0000";
          const key = { path: path, val_whoid: prefix + whoid };
          client.send(new DeleteCommand({ TableName: pathTable.TableName, Key: key }))
            .then(() => cb(null))
            .catch(err => cb(err));
        });

        callbacks.push(cb => {
          if (diff.deleted) return cb(null);
          const prefix = _.isEqual(diff.cur, {}) ? "" : diff.cur + "\u0000";
          const item = { _whoid: whoid, path: path, val_whoid: prefix + whoid };
          client.send(new PutCommand({ TableName: pathTable.TableName, Item: item }))
            .then(() => cb(null))
            .catch(err => cb(err));
        });

        callbacks.push(cb => {
          let item = {
            _whoid: whoid,
            path_time: path + ".\u0000" + now.getTime(),
            path: path,
            date: now.toString(),
          };
          item = _.assign(item, diff);
          client.send(new PutCommand({ TableName: histTable.TableName, Item: item }))
            .then(() => cb(null))
            .catch(err => cb(err));
        });
      });

      async.parallel(callbacks, (err) => {
        if (err) return cb(err);
        cb(null, cur);
      });
    },
  };
}

module.exports = function (
  endpoint,
  region,
  accessId,
  secretKey,
  tableNameSuffix,
  readWriteCapacity,
) {
  objTable.TableName += tableNameSuffix || "";
  pathTable.TableName += tableNameSuffix || "";
  histTable.TableName += tableNameSuffix || "";

  if (!Number.isInteger(readWriteCapacity)) {
    throw "Invalid dynamo read/write capacity: " + readWriteCapacity;
  }

  objTable.ProvisionedThroughput = {
    ReadCapacityUnits: readWriteCapacity,
    WriteCapacityUnits: readWriteCapacity,
  };
  pathTable.ProvisionedThroughput = {
    ReadCapacityUnits: readWriteCapacity,
    WriteCapacityUnits: readWriteCapacity,
  };
  histTable.ProvisionedThroughput = {
    ReadCapacityUnits: readWriteCapacity,
    WriteCapacityUnits: readWriteCapacity,
  };

  let pending = [];
  let waitForCreate = (thunk) => {
    pending.push(thunk);
  };

  createTablesIfNeeded(endpoint, region, accessId, secretKey, (err) => {
    if (err) {
      throw err;
    }

    let working = createWorkingExport(endpoint, region, accessId, secretKey);
    Object.keys(working).forEach((key) => {
      loading[key] = working[key];
    });
    log.info("dynamo db ready");

    waitForCreate = null;
    pending.forEach((thunk) => {
      thunk();
    });
    pending = null;
  });

  let loading = {
    all(cb) {
      waitForCreate(() => loading.all(cb));
    },
    has(key, cb) {
      waitForCreate(() => loading.has(key, cb));
    },
    by(key, val, cb) {
      waitForCreate(() => loading.by(key, val, cb));
    },
    history(whoid, path, cb) {
      waitForCreate(() => loading.history(whoid, path, cb));
    },
    put(cur, diffs, cb) {
      waitForCreate(() => loading.put(cur, diffs, cb));
    },
  };

  return loading;
};
