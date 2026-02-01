const https = require("https")
const axios = require("axios")
const isJson = require("is-json")

const instance = axios.create({
  httpsAgent: new https.Agent({
    rejectUnauthorized: false
  })
});

async function request(url, config = {}) {
  try {
    const res = await instance.request({
      timeout: 5000,
      ...config,
      headers: {
        ...config.headers,
        "User-Agent": "Mozilla/5.0 (Linux 64x; Linux; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36"
      },
      url: String(url || ""),
    })
    return {
      isJson: typeof res.data === "object"? true : isJson(res.data),
      isError: false,
      statusText: res.statusText,
      headers: res.headers,
      status: res.status,
      data: res.data,
    }
  } catch (e) {
    const reserr = e.response
    return {
      isJson: !!reserr ? isJson(reserr?.data) : false,
      isError: true,
      statusText: reserr?.statusText || "Unknowing",
      headers: reserr?.headers || {},
      status: reserr?.status || -3,
      data: reserr?.data || {},
    }
  }
}

module.exports = request