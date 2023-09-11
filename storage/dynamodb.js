const aws = require("aws-sdk");
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
  dynamodb.describeTable({ TableName: table.TableName }, (err, data) => {
    if (!err) {
      return cb(checkSchema(table, data.Table));
    }

    log.warn("creating table " + table.TableName);
    dynamodb.createTable(table, cb);
  });
}

function createTablesIfNeeded(endpoint, region, accessId, secretKey, cb) {
  let dynamodb = new aws.DynamoDB({
    endpoint: endpoint,
    region: region,
    accessKeyId: accessId,
    secretAccessKey: secretKey,
  });

  async.parallel(
    [
      (cb) => createTableIfNeeded(dynamodb, objTable, cb),
      (cb) => createTableIfNeeded(dynamodb, pathTable, cb),
      (cb) => createTableIfNeeded(dynamodb, histTable, cb),
    ],
    cb,
  );
}

function createWorkingExport(endpoint, region, accessId, secretKey) {
  let client = new aws.DynamoDB.DocumentClient({
    endpoint: endpoint,
    region: region,
    accessKeyId: accessId,
    secretAccessKey: secretKey,
  });

  let toObj = (obj, cb) => {
    var params = { TableName: objTable.TableName, Key: { _whoid: obj._whoid } };
    client.get(params, (err, data) => {
      cb(err, data?.Item);
    });
  };

  return {
    all(cb) {
      client.scan({ TableName: objTable.TableName }, (err, scan) => {
        if (err) {
          return cb(err);
        }

        cb(null, scan?.Items);
      });
    },
    has(key, cb) {
      var params = {
        TableName: pathTable.TableName,
        KeyConditionExpression: "#path = :key",
        ExpressionAttributeNames: { "#path": "path" },
        ExpressionAttributeValues: { ":key": key },
      };

      client.query(params, function (err, data) {
        if (err) {
          return cb(err);
        }

        async.map(data?.Items, toObj, cb);
      });
    },
    by(key, val, cb) {
      var params = {
        TableName: pathTable.TableName,
        KeyConditionExpression: "#path = :key and begins_with(val_whoid, :val)",
        ExpressionAttributeNames: { "#path": "path" },
        ExpressionAttributeValues: { ":key": key, ":val": val + "\u0000" },
      };

      client.query(params, function (err, data) {
        if (err) {
          return cb(err);
        }

        async.map(data?.Items, toObj, cb);
      });
    },
    history(whoid, path, cb) {
      var params = {
        TableName: histTable.TableName,
        KeyConditionExpression: "#whoid = :whoid",
        ExpressionAttributeNames: { "#whoid": "_whoid" },
        ExpressionAttributeValues: { ":whoid": whoid },
      };

      if (path) {
        params.KeyConditionExpression += " and begins_with(path_time, :path)";
        params.ExpressionAttributeValues[":path"] = path + ".";
      }

      client.query(params, function (err, data) {
        if (err) {
          return cb(err);
        }

        let results = data?.Items.map((d) => _.omit(d, ["path_time"]));
        cb(null, results);
      });
    },
    put(cur, diffs, cb) {
      let callbacks = [];

      let whoid = cur._whoid;
      let now = new Date();

      callbacks.push((cb) => {
        client.put({ TableName: objTable.TableName, Item: cur }, cb);
      });
      Object.keys(diffs).forEach((path) => {
        let diff = diffs[path];

        callbacks.push(
          (cb) => {
            if (diff.created) {
              return cb();
            }

            // value of {} means path has sub object
            let prefix = _.isEqual(diff.prev, {}) ? "" : diff.prev + "\u0000";
            let key = { path: path, val_whoid: prefix + whoid };
            client.delete({ TableName: pathTable.TableName, Key: key }, cb);
          },
          (cb) => {
            if (diff.deleted) {
              return cb();
            }

            // value of {} means path has sub object
            let prefix = _.isEqual(diff.cur, {}) ? "" : diff.cur + "\u0000";
            let item = { _whoid: whoid, path: path, val_whoid: prefix + whoid };
            client.put({ TableName: pathTable.TableName, Item: item }, cb);
          },
          (cb) => {
            let item = {
              _whoid: whoid,
              path_time: path + ".\u0000" + now.getTime(),
              path: path,
              date: now.toString(),
            };
            item = _.assign(item, diff);

            client.put({ TableName: histTable.TableName, Item: item }, cb);
          },
        );
      });

      // TODO look into making all these writes atomic
      async.parallel(callbacks, (err) => {
        if (err) {
          return cb(err);
        }

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
