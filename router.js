const bee = require("beeline");
const _ = require("lodash");

function omitId(objs) {
  if (Array.isArray(objs)) {
    return objs.map(omitId);
  } else if (_.isPlainObject(objs)) {
    return _.mapValues(_.omit(objs, "_whoid"), omitId);
  } else {
    return objs;
  }
}

function genericResponder(res) {
  return (err, data) => {
    if (err) {
      if (err.isUserError) {
        return res.status(400).json({ error: err.toString() });
      }

      return res.status(500).json({ error: err.toString() });
    }
    if (data == null) {
      return res.status(404).json({ error: "not-found" });
    }
    if (data != null && _.isEmpty(data)) {
      return res.status(404).json(data);
    }

    return res.json(omitId(data));
  };
}

module.exports = function (db) {
  return bee.route({
    "/health": (req, res) => {
      res.json({ ok: true });
    },
    "/alias /all /list": {
      GET: (req, res) => {
        db.all(genericResponder(res));
      },
    },
    "/alias/`key` /list/`key`": {
      GET: (req, res) => {
        let key = req.params["key"];

        db.has(key, genericResponder(res));
      },
    },
    "/alias/`key`/`value`": {
      GET: (req, res) => {
        let key = req.params["key"];
        let value = req.params["value"];

        db.one(key, value, genericResponder(res));
      },
      "POST PUT": (req, res) => {
        let author = req.headers["x-wiw-author"];
        if (!author) {
          return res.status(400).json({ error: "no X-WIW-Author header" });
        }

        let key = req.params["key"];
        let value = req.params["value"];

        db.put(author, key, value, req.body, genericResponder(res));
      },
    },
    "/list/`key`/`value`": {
      GET: (req, res) => {
        let key = req.params["key"];
        let value = req.params["value"];

        db.by(key, value, genericResponder(res));
      },
    },
    "/alias/`key`/`value`/data/`path...`": {
      GET: (req, res) => {
        let key = req.params["key"];
        let value = req.params["value"];

        db.one(key, value, (err, data) => {
          if (err) {
            return res.status(500).json({ error: err.toString() });
          }
          if (!data) {
            return res.status(404).json({ error: "not-found" });
          }

          let path = req.params["path"].split("/");
          data = _.get(omitId(data), path);

          if (data == null) {
            return res.status(404).json({ error: "not-found" });
          }

          return res.json(data);
        });
      },
      "POST PUT": (req, res) => {
        let author = req.headers["x-wiw-author"];
        if (!author) {
          return res.status(400).json({ error: "no X-WIW-Author header" });
        }

        let key = req.params["key"];
        let value = req.params["value"];

        let path = req.params["path"].split("/");
        let data = _.set({}, path, req.body);

        db.put(author, key, value, data, genericResponder(res));
      },
    },
    "/list/`key`/`value`/data/`path...`": {
      GET: (req, res) => {
        let key = req.params["key"];
        let value = req.params["value"];

        db.by(key, value, (err, data) => {
          if (err) {
            return res.status(500).json({ error: err.toString() });
          }
          if (_.isEmpty(data)) {
            return res.status(404).json(data);
          }

          data = omitId(data);

          let path = req.params["path"].replace("/", ".");
          return res.json(_.flatten(_.map(data, path)));
        });
      },
    },
    "/alias/`key`/`value`/history/`path...`": {
      GET: (req, res) => {
        let key = req.params["key"];
        let value = req.params["value"];
        let path = req.params["path"].replace("/", ".");

        db.history(key, value, path, genericResponder(res));
      },
    },
  });
};
