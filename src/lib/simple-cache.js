let memoryCache = {} // Ram Cache

function set(key = "", value) {
  memoryCache[String(key)] = value
  return true
}

function get(key = "") {
  return memoryCache[String(key)]
}

function del(key = "") {
  delete memoryCache[String(key)]
  return true
}

module.exports = {
  set, del, get
}