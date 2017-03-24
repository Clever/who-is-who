const emitter = require("events").EventEmitter;
const assert = require("assert");

const _ = require("lodash");
const mocks = require("node-mocks-http");
const async = require("async");

let endpoint = process.env.AWS_DYNAMO_ENDPOINT;
let storage = require("../storage/dynamodb")(
  endpoint,
  "us-west-1",
  "test",
  "test",
  "test",
  1
);

const db = require("../db")(storage);
const router = require("../router")(db);

let mockData = [
  {email: "1@mail.com", uniq: 1, hi: 1, a: "a", deep: {a: 1}},
  {email: "2@mail.com", uniq: 2, hi: 1, a: "ab", deep: {b: 1}},
  {email: "3@mail.com", uniq: 3, hi: 2, a: "abc", deep: {c: 1}},
  {email: "4@mail.com", uniq: 4, bye: 1, a: "abcd", deeper: {d: {e: 1}}}
];
let dbPopulated = false;
exports.setUp = function(done) {
  if (dbPopulated) {
    return done();
  }

  async.each(
    mockData,
    (data, cb) => db.put("init", "email", data.email, data, cb),
    err => {
      if (err) {
        throw err;
      }

      dbPopulated = true;
      done();
    }
  );
};
exports.tearDown = function(done) {
  done();
};

function mockSend(opt, cb) {
  let req = mocks.createRequest(opt);
  let res = mocks.createResponse({eventEmitter: emitter});

  res.on("end", () => {
    assert.ok(res._isJSON);

    let data = JSON.parse(res._getData());
    cb(res, data);
  });

  router(req, res);

  return req;
}

function mockGET(url, cb) {
  return mockSend({url: url, method: "GET"}, cb);
}

function mockPOST(url, body, cb) {
  return mockSend(
    {
      url: url,
      body: body,
      method: "POST",
      headers: {"X-WIW-Author": "mock-post"}
    },
    cb
  );
}

exports["/alias, /all, /list"] = function(test) {
  test.expect(2);

  mockGET("/all", (res, data) => {
    test.equal(res.statusCode, 200);

    data.sort((a, b) => a.uniq - b.uniq);
    test.deepEqual(data, mockData);

    test.done();
  });
};

exports["/alias/`key`, /list/`key`"] = {
  "multi-match": test => {
    test.expect(2);

    mockGET("/alias/hi", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 3);
      test.done();
    });
  },
  "no pre-fix for deep keys": test => {
    test.expect(2);

    mockGET("/list/de", (res, data) => {
      test.equal(res.statusCode, 404);
      test.deepEqual(data, {});
      test.done();
    });
  },
  "no pre-fix for values": test => {
    test.expect(2);

    mockGET("/alias/a/a", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.uniq, 1);
      test.done();
    });
  },
  "deep key": test => {
    test.expect(3);

    mockGET("/list/deep.a", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 1);
      test.equal(data[0].uniq, 1);
      test.done();
    });
  },
  "deeper key": test => {
    test.expect(3);

    mockGET("/alias/deeper.d.e", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 1);
      test.equal(data[0].uniq, 4);
      test.done();
    });
  },
  "no-matches": test => {
    test.expect(2);

    mockGET("/list/no_matches", (res, data) => {
      test.equal(res.statusCode, 404);
      test.equal(data.length, 0);
      test.done();
    });
  }
};

exports["/alias/`key`/`value`"] = {
  uniq: test => {
    test.expect(2);

    mockGET("/alias/hi/2", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.uniq, 3);
      test.done();
    });
  },
  "/data/`path...` uniq": test => {
    test.expect(2);

    mockGET("/alias/hi/2/data/deep", (res, data) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {c: 1});
      test.done();
    });
  },
  "multi-match error": test => {
    test.expect(2);

    mockGET("/alias/hi/1", (res, data) => {
      test.equal(res.statusCode, 500);
      test.ok(data["error"]);
      test.done();
    });
  },
  "/data/`path...` multi-match error": test => {
    test.expect(2);

    mockGET("/alias/hi/1/data/deep", (res, data) => {
      test.equal(res.statusCode, 500);
      test.ok(data["error"]);
      test.done();
    });
  },
  "no-matches": test => {
    test.expect(2);

    mockGET("/alias/hi/3", (res, data) => {
      test.equal(res.statusCode, 404);
      test.ok(data["error"]);
      test.done();
    });
  },
  "/data/`path...` no-matches": test => {
    test.expect(2);

    mockGET("/alias/hi/3/data/deep", (res, data) => {
      test.equal(res.statusCode, 404);
      test.ok(data["error"]);
      test.done();
    });
  },
  "POST, PUT with no author": test => {
    test.expect(2);

    let opts = {
      url: "/alias/bye/2",
      body: {cheggit: "yo", email: "5@mail.com"},
      method: "POST",
      headers: {}
    };
    mockSend(opts, (res, data) => {
      test.equal(res.statusCode, 400);
      test.ok(data.error);
      test.done();
    });
  },
  "POST, PUT with no author deep data": test => {
    test.expect(2);

    let opts = {
      url: "/alias/email/5@email.com/data/deep",
      body: {cheggit: "yo"},
      method: "POST",
      headers: {}
    };
    mockSend(opts, (res, data) => {
      test.equal(res.statusCode, 400);
      test.ok(data.error);
      test.done();
    });
  },
  "POST, PUT with no email": test => {
    test.expect(2);

    let opts = {
      url: "/alias/foo/bar",
      body: {cheggit: "yo"},
      method: "POST",
      headers: {}
    };
    mockSend(opts, (res, data) => {
      test.equal(res.statusCode, 400);
      test.ok(data.error);
      test.done();
    });
  },
  "POST, PUT keys with dots": test => {
    test.expect(2);

    let data = {"cheg.git": "yo", email: "7@mail.com"};
    mockPOST("/alias/bye/3", data, (res, data) => {
      test.equal(res.statusCode, 400);
      test.ok(data.error);
      test.done();
    });
  },
  "POST, PUT with empty values": test => {
    test.expect(2);

    let data = {cheggit: "", email: "5@mail.com", num: 0, bool: false};
    mockPOST("/alias/bye/4", data, (res, data) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {bye: 4, email: "5@mail.com", num: 0, bool: false});
      test.done();
    });
  },
  "POST, PUT with deep empty values": test => {
    test.expect(2);

    let data = {
      deepish: {cheggit: "", num: 0, bool: false},
      email: "325@mail.com"
    };
    mockPOST("/alias/bye/5", data, (res, data) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {
        bye: 5,
        email: "325@mail.com",
        deepish: {num: 0, bool: false}
      });
      test.done();
    });
  },
  "POST, PUT with null values": test => {
    test.expect(2);

    let data = {cheggit: null, email: "538@mail.com"};
    mockPOST("/alias/bye/6", data, (res, data) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {bye: 6, email: "538@mail.com"});
      test.done();
    });
  },
  "POST, PUT with deep null values": test => {
    test.expect(2);

    let data = {deepish: {cheggit: null}, email: "0935@mail.com"};
    mockPOST("/alias/bye/7", data, (res, data) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {bye: 7, email: "0935@mail.com", deepish: {}});
      test.done();
    });
  },
  "POST, PUT new value": test => {
    test.expect(7);

    mockPOST("/alias/bye/2", {cheggit: "yo", email: "745@mail.com"}, (
      res,
      data
    ) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {bye: 2, cheggit: "yo", email: "745@mail.com"});

      mockGET("/alias/bye/2/history/", (res, data) => {
        test.equal(res.statusCode, 200);
        test.equal(Object.keys(data).length, 3);
        test.deepEqual(data["bye"].map(o => _.omit(o, "date")), [
          {created: true, cur: 2, author: "mock-post"}
        ]);
        test.deepEqual(data["cheggit"].map(o => _.omit(o, "date")), [
          {created: true, cur: "yo", author: "mock-post"}
        ]);
        test.deepEqual(data["email"].map(o => _.omit(o, "date")), [
          {created: true, cur: "745@mail.com", author: "mock-post"}
        ]);
        test.done();
      });
    });
  },
  "POST, PUT new value deep data": test => {
    test.expect(7);

    mockPOST("/alias/email/6@mail.com/data/greets", {cheggit: "yo"}, (
      res,
      data
    ) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {email: "6@mail.com", greets: {cheggit: "yo"}});

      mockGET("/alias/email/6@mail.com/history/", (res, data) => {
        test.equal(res.statusCode, 200);
        test.equal(Object.keys(data).length, 3);
        test.deepEqual(data["greets"].map(o => _.omit(o, "date")), [
          {created: true, cur: {}, author: "mock-post"}
        ]);
        test.deepEqual(data["email"].map(o => _.omit(o, "date")), [
          {created: true, cur: "6@mail.com", author: "mock-post"}
        ]);
        test.deepEqual(data["greets.cheggit"].map(o => _.omit(o, "date")), [
          {created: true, cur: "yo", author: "mock-post"}
        ]);
        test.done();
      });
    });
  },
  "POST, PUT objects with the same email address are merged": test => {
    test.expect(11);

    mockPOST("/alias/email/peep@peep.com", {cheggit: "yo"}, (res, data) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {cheggit: "yo", email: "peep@peep.com"});


      mockPOST("/alias/slack/peep", {orto: "next", "email": "peep@peep.com"}, (res, data) => {
        test.equal(res.statusCode, 200);
        test.deepEqual(data, {
          cheggit: "yo", orto: "next", email: "peep@peep.com", slack: "peep"
        });

        mockGET("/alias/email/peep@peep.com", (res, data) => {
          test.equal(res.statusCode, 200);
          test.deepEqual(data, {
            cheggit: "yo", orto: "next", email: "peep@peep.com", slack: "peep"
          });

          mockGET("/alias/email/peep@peep.com/history/", (res, data) => {
            test.equal(Object.keys(data).length, 4);
            test.deepEqual(data["orto"].map(o => _.omit(o, "date")), [
              {created: true, cur: "next", author: "mock-post"}
            ]);
            test.deepEqual(data["cheggit"].map(o => _.omit(o, "date")), [
              {created: true, cur: "yo", author: "mock-post"}
            ]);
            test.deepEqual(data["email"].map(o => _.omit(o, "date")), [
              {created: true, cur: "peep@peep.com", author: "mock-post"}
            ]);
            test.deepEqual(data["slack"].map(o => _.omit(o, "date")), [
              {created: true, cur: "peep", author: "mock-post"}
            ]);
            test.done();
          });
        });
      });
    });
  },
  "POST, PUT editting email address works": test => {
    test.expect(7);

    mockPOST("/alias/email/poop@poop.com", {slack: "poop"}, (res, data) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {slack: "poop", email: "poop@poop.com"});

      mockPOST("/alias/slack/poop", {"email": "poop2@poop.com"}, (res, data) => {
        test.equal(res.statusCode, 200);
        test.deepEqual(data, { email: "poop2@poop.com", slack: "poop" });

        mockGET("/alias/email/poop@poop.com", (res, data) => {
          test.equal(res.statusCode, 404);

          mockGET("/alias/email/poop2@poop.com", (res, data) => {
            test.equal(res.statusCode, 200);
            test.deepEqual(data, { email: "poop2@poop.com", slack: "poop" });

            test.done();
          });
        });
      });
    });
  },
  "POST, PUT existing value": test => {
    test.expect(7);

    mockPOST("/alias/uniq/1", {hi: 2}, (res, data) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {
        email: "1@mail.com",
        uniq: 1,
        hi: 2,
        a: "a",
        deep: {a: 1}
      });

      mockGET("/alias/uniq/1/history/hi", (res, data) => {
        test.equal(res.statusCode, 200);
        test.equal(Object.keys(data).length, 1);
        test.deepEqual(data["hi"].map(o => _.omit(o, "date")), [
          {prev: 1, cur: 2, author: "mock-post"},
          {created: true, cur: 1, author: "init"}
        ]);

        mockPOST("/alias/uniq/1", {hi: 1}, (res, data) => {
          test.equal(res.statusCode, 200);
          test.deepEqual(data, {
            // Reset mock data
            email: "1@mail.com",
            uniq: 1,
            hi: 1,
            a: "a",
            deep: {a: 1}
          });
          test.done();
        });
      });
    });
  },
  "POST, PUT existing value deep data": test => {
    test.expect(9);

    mockPOST("/alias/uniq/1/data/deep", {a: 2}, (res, data) => {
      test.equal(res.statusCode, 200);
      test.deepEqual(data, {
        email: "1@mail.com",
        uniq: 1,
        hi: 1,
        deep: {a: 2},
        a: "a"
      });

      mockGET("/alias/uniq/1/history/deep/a", (res, data) => {
        test.equal(res.statusCode, 200);
        test.equal(Object.keys(data).length, 1);
        test.deepEqual(data["deep.a"].map(o => _.omit(o, "date")), [
          {prev: 1, cur: 2, author: "mock-post"},
          {created: true, cur: 1, author: "init"}
        ]);

        mockPOST("/alias/uniq/1/data/deep", {a: 1}, (res, data) => {
          test.equal(res.statusCode, 200);
          test.deepEqual(data, {
            // Reset mock data
            email: "1@mail.com",
            uniq: 1,
            hi: 1,
            a: "a",
            deep: {a: 1}
          });

          mockGET("/alias/uniq/1/history/de", (res, data) => {
            test.equal(res.statusCode, 404);
            test.deepEqual(data, {});
            test.done();
          });
        });
      });
    });
  },
  "POST, PUT deleting deep object": test => {
    test.expect(6);

    mockPOST("/alias/email/8@email.com", {a: 2, b: 3, c: {d: 6}}, (
      res,
      data
    ) => {
      test.equal(res.statusCode, 200);
      test.ok(data.c);

      mockPOST("/alias/email/8@email.com", {a: 2, b: 3, c: null}, (
        res,
        data
      ) => {
        test.equal(res.statusCode, 200);
        test.ok(!data.c);

        mockGET("/list/c", (res, data) => {
          test.equal(res.statusCode, 404);
          test.equal(0, data.length);

          test.done();
        });
      });
    });
  }
};

exports["/list/`key`"] = {
  "no vals": test => {
    test.expect(2);

    mockGET("/list/nudting", (res, data) => {
      test.equal(res.statusCode, 404);
      test.equal(data.length, 0);
      test.done();
    });
  },
  "shallow key": test => {
    test.expect(3);

    mockGET("/list/hi", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 3);
      test.deepEqual(data.map(d => d.uniq).sort(), [1, 2, 3]);
      test.done();
    });
  },
  "deep key": test => {
    test.expect(3);

    mockGET("/list/deep", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 3);
      test.deepEqual(data.map(d => d.uniq).sort(), [1, 2, 3]);
      test.done();
    });
  }
};

exports["/list/`key`/`value`"] = {
  "no vals": test => {
    test.expect(2);

    mockGET("/list/nudting/hello", (res, data) => {
      test.equal(res.statusCode, 404);
      test.equal(data.length, 0);
      test.done();
    });
  },
  "shallow key": test => {
    test.expect(3);

    mockGET("/list/hi/1", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 2);
      test.deepEqual(data.map(d => d.uniq).sort(), [1, 2]);
      test.done();
    });
  },
  "single key": test => {
    test.expect(3);

    mockGET("/list/hi/2", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 1);
      test.deepEqual(data[0].uniq, 3);
      test.done();
    });
  }
};

exports["/list/`key`/`value`/data/`path...`"] = {
  uniq(test) {
    test.expect(3);

    mockGET("/list/uniq/1", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 1);
      test.equal(data[0].uniq, 1);
      test.done();
    });
  },
  path(test) {
    test.expect(3);

    mockGET("/list/deeper.d.e/1", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 1);
      test.equal(data[0].uniq, 4);
      test.done();
    });
  },
  multi(test) {
    test.expect(2);

    mockGET("/list/hi/1", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 2);
      test.done();
    });
  },
  "/data/`path...` multi": test => {
    test.expect(2);

    mockGET("/list/hi/1/data/deep", (res, data) => {
      test.equal(res.statusCode, 200);
      test.equal(data.length, 2);
      test.done();
    });
  },
  "no-matches": test => {
    test.expect(2);

    mockGET("/list/hi/3", (res, data) => {
      test.equal(res.statusCode, 404);
      test.equal(data.length, 0);
      test.done();
    });
  },
  "/data/`path...` no-matches": test => {
    test.expect(2);

    mockGET("/list/hi/3/data/deep", (res, data) => {
      test.equal(res.statusCode, 404);
      test.equal(data.length, 0);
      test.done();
    });
  }
};
