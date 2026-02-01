const ymalSys = require("yaml")
const fs = require("fs")

const defaultConfigPath = "./config/proxmox-node.yml"

function UpdateConfig() {
  console.log("[Config]: Load file : "+defaultConfigPath)
  let monitoringNode = []

  if(!!fs.existsSync(defaultConfigPath) && !!fs.lstatSync(defaultConfigPath)?.isFile()) {
    const parserData = ymalSys.parse(
      fs.readFileSync(defaultConfigPath, "utf-8")
    )
    monitoringNode = (parserData?.monitoring||[])
  }
  return monitoringNode
}

module.exports = UpdateConfig