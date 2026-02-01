const express = require("express")
const cors = require("cors")

const metricsBuilder = require("./lib/metrics-builder")
const configReader = require("./lib/config-reader")
const requestData = require("./lib/request-data")

const app = express()
app.use(cors())

app.get("/metrics", async (req, res) => {
  // Load Config
  const loadConfig = configReader()
  const prepareConfig = loadConfig||[]
  // Fetching Data
  const fetchingData = prepareConfig.map((item) => {
    return requestData.callAllRequest(item)
  })
  const dataScrapping = await Promise.all(fetchingData)
  // Building Metrics
  metricsBuilder.buildingMetrics(dataScrapping)
  // Response
  res.set("Content-Type", metricsBuilder.registry.contentType)
  res.send(await metricsBuilder.registry.metrics())
})

app.listen(8007, () => {
  console.log("Server running on port 8007\nURL: http://localhost:8007/metrics")
})
