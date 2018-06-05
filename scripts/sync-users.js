const fs = require("fs");
const readline = require("readline");
const google = require("googleapis");
const googleAuth = require("google-auth-library");
const exec = require("child_process").exec;

// Heavily inspired by https://developers.google.com/admin-sdk/directory/v1/quickstart/nodejs

const SCOPES = ["https://www.googleapis.com/auth/admin.directory.user"];
const TOKEN_DIR = ".";
const TOKEN_PATH = "token.json";

let endpoint = "https://dynamodb.us-west-1.amazonaws.com/";
let region = "us-west-1";
let accessKeyId = process.env.AWS_ACCESS_KEY_ID;
let secretAccessKey = process.env.AWS_SECRET_ACCESS_KEY;

const storage = require("./../storage/dynamodb")(
  endpoint,
  region,
  accessKeyId,
  secretAccessKey,
  "",
  5
);
const db = require("./../db")(storage);

const cmd = "ark secrets read production.who-is-who google-client-key --no-upgrade";

exec(cmd, function(err, content) {
  if (err) {
    console.log("Error reading secret: " + err);
    return;
  }
  authorize(JSON.parse(content), listUsers);
});

function authorize(credentials, callback) {
  const clientSecret = credentials.installed.client_secret;
  const clientId = credentials.installed.client_id;
  const redirectUrl = credentials.installed.redirect_uris[0];
  const auth = new googleAuth();
  const oauth2Client = new auth.OAuth2(clientId, clientSecret, redirectUrl);

  // Check if we have previously stored a token.
  fs.readFile(TOKEN_PATH, function(err, token) {
    if (err) {
      console.log("Getting new token");
      getNewToken(oauth2Client, callback);
    } else {
      oauth2Client.credentials = JSON.parse(token);
      callback(oauth2Client);
    }
  });
}

function getNewToken(oauth2Client, callback) {
  const authUrl = oauth2Client.generateAuthUrl({
    access_type: "offline",
    scope: SCOPES
  });
  console.log("Authorize this app by visiting this url: ", authUrl);
  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });
  rl.question("Enter the code from that page here: ", function(code) {
    rl.close();
    oauth2Client.getToken(code, function(err, token) {
      if (err) {
        console.log("Error while trying to retrieve access token", err);
        return;
      }
      oauth2Client.credentials = token;
      storeToken(token);
      callback(oauth2Client);
    });
  });
}

function storeToken(token) {
  try {
    fs.mkdirSync(TOKEN_DIR);
  } catch (err) {
    if (err.code != "EEXIST") {
      throw err;
    }
  }
  fs.writeFile(TOKEN_PATH, JSON.stringify(token));
  console.log("Token stored to " + TOKEN_PATH);
}

function listUsers(auth) {
  const service = google.admin("directory_v1");
  service.users.list(
    {
      auth: auth,
      customer: "my_customer",
      maxResults: 200,
      orderBy: "email"
    },
    (err, response) => {
      if (err) {
        console.log("The API returned an error: " + err);
        return;
      }
      const users = response.users;
      if (users.length == 0) {
        console.log("No users in the domain.");
      } else {
        db.all((err, data) => {
          const googleEmails = users
            .filter(x => x.orgUnitPath == "/FTEs")
            .map(u => u.primaryEmail);
          let activeWhoIsWho = data.filter(u => u.active);
          let inSync = true;
          for (const googleEmail of googleEmails) {
            const whoIsWho = data.filter(u => u.email == googleEmail);
            let employee = {};
            if (whoIsWho.length != 1) {
              console.log("Missing who is who for: " + googleEmail);
              employee = {email: googleEmail};
            } else {
              employee = whoIsWho[0];
            }
            if (employee.active) {
              activeWhoIsWho = activeWhoIsWho.filter(
                u => u.email !== googleEmail
              );
            } else {
              inSync = false;
              console.log("Not marked active, should be: " + googleEmail);
              employee.active = true;
              db.put(
                "sync-users-script",
                "email",
                employee.email,
                employee,
                err => {
                  if (err) {
                    console.log(
                      "Error updating user (" + employee.email + "): " + err
                    );
                  }
                }
              );
            }
          }
          for (const employee of activeWhoIsWho) {
            inSync = false;
            console.log("Marked active, should not be: " + employee.email);
            employee.active = false;
            db.put(
              "sync-users-script",
              "email",
              employee.email,
              employee,
              err => {
                if (err) {
                  console.log(
                    "Error updating user (" + employee.email + "): " + err
                  );
                }
              }
            );
          }
          if (inSync) {
            console.log("You're in sync! Bye bye bye.");
          }
        });
      }
    }
  );
}
