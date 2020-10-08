const express = require("express");
const bodyParser = require("body-parser");

const kv = require("kayvee");

const log = new kv.logger("who-is-who");

let port = process.env.PORT || "80";

let endpoint = process.env.AWS_DYNAMO_ENDPOINT || "https://dynamodb.us-west-1.amazonaws.com/";
let region = process.env.AWS_DYNAMO_REGION || "us-west-1";
let accessKeyId = process.env.AWS_ACCESS_KEY_ID;
let secretAccessKey = process.env.AWS_SECRET_ACCESS_KEY;
let tableNameSuffix = process.env.TABLE_NAME_SUFFIX;
let dynamoReadWriteCapacity = parseInt(process.env.DYNAMO_READ_WRITE_CAPACITY);

const storage = require("./storage/dynamodb")(
  endpoint,
  region,
  accessKeyId,
  secretAccessKey,
  tableNameSuffix,
  dynamoReadWriteCapacity,
);
const db = require("./db")(storage);
const router = require("./router")(db);

let app = express();

app.use(kv.middleware({ source: "who-is-who" }));
app.use(bodyParser.json());

app.use(router).listen(port);

log.infoD("server ready", { "listening on": port });
