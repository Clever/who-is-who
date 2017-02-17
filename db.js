const _ = require("lodash");
const uuid = require("uuid");

function UserError(str) {
	let err = new Error(str);
	err.isUserError = true;

	return err;
}

function parseIntIfNeeded(n) {
	if (!isNaN(parseFloat(n)) && isFinite(n)) {
		return parseFloat(n);
	} else {
		return n;
	}
}

function toPaths(obj) {
	let paths = {};
	let node, nodes = [{path: "", obj: obj}];

	while ((node = nodes.pop()) != undefined) {
		let obj = node.obj, path = node.path;

		if (!_.isPlainObject(obj)) {
			paths[path] = obj;
			continue;
		} else {
			paths[path] = {};
		} // Empty object means path contains a sub object

		Object.keys(obj).forEach(key => {
			let next = {
				path: !path ? key : path + "." + key,
				obj: obj[key]
			};
			nodes.push(next);
		});
	}

	return paths;
}

function generateDiffs(author, prev, cur) {
	let diffs = {};

	let prevPaths = toPaths(prev);
	let curPaths = toPaths(cur);

  // Find deleted paths
	Object.keys(prevPaths).forEach(path => {
		if (path in curPaths) {
			return;
		}

		diffs[path] = {deleted: true, prev: prevPaths[path], author: author};
	});

  // Find changed and created paths
	Object.keys(curPaths).forEach(path => {
		let prev = prevPaths[path];
		let cur = curPaths[path];

		if (_.isEqual(prev, cur)) {
			return;
		}

		if (path in prevPaths) {
			diffs[path] = {prev: prev, cur: cur};
		} else {
			diffs[path] = {created: true, cur: cur};
		}
		diffs[path].author = author;
	});

	return diffs;
}

function removeEmptyValues(obj) {
	obj = _.cloneDeep(obj);

	let node, nodes = [obj];
	while ((node = nodes.pop()) != undefined) {
		Object.keys(node).forEach(key => {
			if (node[key] === "" || _.isNil(node[key])) {
				delete node[key];
			}
		});

		let children = _.filter(node, _.isPlainObject);
		nodes.push(...children);
	}
	return obj;
}

module.exports = function(storage) {
	return {
		all(cb) {
			storage.all(cb);
		},
		has(key, cb) {
			storage.has(key, cb);
		},
		by(key, val, cb) {
			val = parseIntIfNeeded(val);

			storage.by(key, val, cb);
		},
		one(key, val, cb) {
			val = parseIntIfNeeded(val);

			this.by(key, val, (err, results) => {
				if (err) {
					return cb(err);
				}
				if (results.length === 0) {
					return cb(null, null);
				}
				if (results.length !== 1) {
					return cb(new Error("More than one value found."));
				}

				return cb(null, results[0]);
			});
		},
		history(key, val, path, cb) {
			this.one(key, val, (err, obj) => {
				if (err) {
					return cb(err);
				}
				if (obj == null) {
					return cb(null, null);
				}

				storage.history(obj._whoid, path, (err, diffs) => {
					if (err) {
						return cb(err);
					}

					let hist = {};
					diffs.forEach(diff => {
						let d = _.omit(diff, "path");

						hist[diff.path] = hist[diff.path] || [];
						d.date = new Date(d.date);
						hist[diff.path].push(d);
					});

					Object.keys(hist).forEach(key => {
						hist[key].sort((a, b) => b.date - a.date);
					});

					cb(null, hist);
				});
			});
		},
		put(author, key, val, body, cb) {
			val = parseIntIfNeeded(val);

			this.one(key, val, (err, obj) => {
				if (err) {
					return cb(err);
				}

				let prev = obj || {_whoid: uuid()};
				let cur = _.defaultsDeep({}, body, prev);
				cur = _.set(cur, key, val); // Here to ensure new values have the correct index

				cur = removeEmptyValues(cur); // Don't save empty values

				if (!cur.email) {
					return cb(UserError("Can't save.  No email address found."));
				}
				if (Object.keys(cur).some(k => k.indexOf(".") !== -1)) {
					return cb(UserError("Key with dots ('.') aren't supported"));
				}

				let diffs = generateDiffs(author, prev, cur);
				storage.put(cur, diffs, cb);
			});
		}
	};
};
