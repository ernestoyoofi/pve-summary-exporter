const request = require("./scrapper")
const simpleCache = require("./simple-cache")

async function GetCrendetials({ host = "", username = "", password = "", realm = "pam", ...etcs } = {}) {
  try {
    const response = await request(`${host}/api2/extjs/access/ticket`, {
      method: "POST",
      headers: {
        "Content-Type": "application/x-www-form-urlencoded"
      },
      data: new URLSearchParams({
        ...etcs,
        username,
        password,
        realm,
      })
    })
    if (response.status !== 200) {
      return {
        isError: true,
        message: String(response.statusText),
        data: {},
      }
    }
    if (!response.isJson) {
      return {
        isError: true,
        message: String(response.statusText),
        data: {},
      }
    }
    console.log("Login Info Status:", response.statusText)
    return {
      isError: false,
      csrftoken: response?.data?.data?.CSRFPreventionToken||"",
      ticket: response?.data?.data?.ticket||"",
    }
  } catch (error) {
    return {
      isError: true,
      message: String(error.message),
      data: {}
    }
  }
}

async function ClusterStatus({ host = "", ticket = "", csrftoken = "" } = {}) {
  try {
    const parseURL = new URL(host)
    const response = await request(`${host}/api2/extjs/cluster/status`, {
      method: "GET",
      headers: {
        "Cookie": `PVEAuthCookie=${String(ticket || "")}`,
        "CSRFPreventionToken": String(csrftoken || ""),
      },
    })
    if (response.status === 401) {
      return {
        isError: true,
        isNeedReAuth: true,
        message: String(response.statusText),
        data: {},
      }
    }
    if (response.status !== 200) {
      return {
        isError: true,
        isNeedReAuth: false,
        message: String(response.statusText),
        data: {},
      }
    }
    if (!response.isJson) {
      return {
        isError: true,
        isNeedReAuth: false,
        message: String(response.statusText),
        data: {},
      }
    }
    const listCluster = (response.data.data||{})
    return {
      isError: false,
      isNeedReAuth: false,
      data: {
        now: listCluster.find((item) => item.ip === parseURL.hostname),
        list: listCluster,
      },
    }
  } catch (error) {
    console.log("Debug Trace ClusterStatus:", error.stack)
    return {
      isError: true,
      isNeedReAuth: false,
      message: String(error),
      data: {}
    }
  }
}

async function ClusterRRDData({ host = "", ticket = "", csrftoken = "", nodeName = "" } = {}) {
  try {
    const response = await request(`${host}/api2/json/nodes/${nodeName}/rrddata?timeframe=hour&cf=AVERAGE`, {
      method: "GET",
      headers: {
        "Cookie": `PVEAuthCookie=${String(ticket || "")}`,
        "CSRFPreventionToken": String(csrftoken || ""),
      },
    })
    if (response.status === 401) {
      return {
        isError: true,
        isNeedReAuth: true,
        message: String(response.statusText),
        data: {},
      }
    }
    if (response.status !== 200) {
      return {
        isError: true,
        isNeedReAuth: false,
        message: String(response.statusText),
        data: {},
      }
    }
    if (!response.isJson) {
      return {
        isError: true,
        isNeedReAuth: false,
        message: String(response.statusText),
        data: {},
      }
    }
    const listGraph = (response.data.data||{})
    const filterWithData = listGraph.filter((item) => typeof (
      item.time && item.netout && item.netin && item.iowait && item.maxcpu && item.cpu && item.memtotal && item.memused && item.swapused && item.swaptotal && item.rootused && item.roottotal && item.loadavg
    ) === "number")
    const shortWithNewData = filterWithData.sort((a, b) => b.time - a.time)[0]
    return {
      isError: false,
      isNeedReAuth: false,
      data: {
        // Memory
        memtotal: shortWithNewData?.memtotal,
        memused: shortWithNewData?.memused,
        memfree: shortWithNewData?.memtotal - shortWithNewData?.memused,
        // Swap
        swaptotal: shortWithNewData?.swaptotal,
        swapused: shortWithNewData?.swapused,
        swapfree: shortWithNewData?.swaptotal - shortWithNewData?.swapused,
        // Root
        rootused: shortWithNewData?.rootused,
        roottotal: shortWithNewData?.roottotal,
        rootfree: shortWithNewData?.roottotal - shortWithNewData?.rootused,
        // Load AVG
        loadavg: shortWithNewData?.loadavg,
        // CPU
        cpu: shortWithNewData?.cpu,
        maxcpu: shortWithNewData?.maxcpu,
        // Input Output
        iowait: shortWithNewData?.iowait,
        // Network
        netin: shortWithNewData?.netin,
        netout: shortWithNewData?.netout,
      },
    }
  } catch (error) {
    console.log("Debug Trace ClusterRRDData:", error.stack)
    return {
      isError: true,
      isNeedReAuth: false,
      message: String(error),
      data: {}
    }
  }
}

async function NodeStatus({ host = "", ticket = "", csrftoken = "", nodeName = "" } = {}) {
  try {
    const response = await request(`${host}/api2/json/nodes/${nodeName}/status`, {
      method: "GET",
      headers: {
        "Cookie": `PVEAuthCookie=${String(ticket || "")}`,
        "CSRFPreventionToken": String(csrftoken || ""),
      },
    })
    if (response.status === 401) {
      return {
        isError: true,
        isNeedReAuth: true,
        message: String(response.statusText),
        data: {},
      }
    }
    if (response.status !== 200) {
      return {
        isError: true,
        isNeedReAuth: false,
        message: String(response.statusText),
        data: {},
      }
    }
    if (!response.isJson) {
      return {
        isError: true,
        isNeedReAuth: false,
        message: String(response.statusText),
        data: {},
      }
    }
    const nodeInfo = (response.data.data||{})
    return {
      isError: false,
      isNeedReAuth: false,
      data: {
        // CPU
        cpu: nodeInfo.cpu,
        // CPU Info
        model: nodeInfo.cpuinfo.model, // String
        cpus: nodeInfo.cpuinfo.cpus,
        cores: nodeInfo.cpuinfo.cores,
        sockets: nodeInfo.cpuinfo.sockets,
        user_hz: nodeInfo.cpuinfo.user_hz,
        mhz: nodeInfo.cpuinfo.mhz, // String
        loadavg1: (nodeInfo.loadavg||[])[0],
        loadavg2: (nodeInfo.loadavg||[])[1],
        loadavg3: (nodeInfo.loadavg||[])[2],
        // SWAP Ram
        swaptotal: nodeInfo.swap.total,
        swapused: nodeInfo.swap.used,
        swapfree: nodeInfo.swap.free,
        // Memory / RAM
        memorytotal: nodeInfo.memory.total,
        memoryused: nodeInfo.memory.used,
        memoryfree: nodeInfo.memory.free,
        // RootFS
        rootfstotal: nodeInfo.rootfs.total,
        rootfsavail: nodeInfo.rootfs.avail,
        rootfsused: nodeInfo.rootfs.used,
        rootfsfree: nodeInfo.rootfs.free,
        // PVE Version
        pveversion: nodeInfo.pveversion, // String
        // Wait
        wait: nodeInfo.wait,
        // Uptime
        uptime: nodeInfo.uptime,
      },
    }
  } catch (error) {
    console.log("Debug Trace NodeStatus:", error.stack)
    return {
      isError: true,
      isNeedReAuth: false,
      message: String(error),
      data: {}
    }
  }
}

async function callAllRequest({ identify, host, ticket } = {}) {
  const timeRequest = new Date().getTime()
  console.log("Reading Info:", { identify, host })
  const keyCache = `${identify}-auth`
  let simpleCacheAuth = simpleCache.get(keyCache)
  if(!simpleCacheAuth) {
    const credentials = await GetCrendetials({
      ...(ticket||{}),
      host: host,
    })
    if(credentials.isError) {
      // Return
      return {
        identify: identify,
        host: host,
        success: false,
        needReAuth: false,
        timeEnd: (new Date().getTime() - timeRequest)
      }
    }
    simpleCacheAuth = credentials
    simpleCache.set(keyCache, credentials)
  }

  const clusterInfo = await ClusterStatus({
    host: host,
    csrftoken: simpleCacheAuth.csrftoken,
    ticket: simpleCacheAuth.ticket
  })
  // Cluster Need Re Auth
  if(clusterInfo.isNeedReAuth) {
    // Return
    return {
      identify: identify,
      host: host,
      success: false,
      needReAuth: true,
      timeEnd: (new Date().getTime() - timeRequest)
    }
  }
  // Cluster Need Error
  if(clusterInfo.isError) {
    // Return
    return {
      identify: identify,
      host: host,
      success: false,
      needReAuth: false,
      timeEnd: (new Date().getTime() - timeRequest)
    }
  }
  const [rrddataInfo, statusInfo] = await Promise.all([
    ClusterRRDData({
      host: host,
      csrftoken: simpleCacheAuth.csrftoken,
      ticket: simpleCacheAuth.ticket,
      nodeName: clusterInfo.data.now.name
    }),
    NodeStatus({
      host: host,
      csrftoken: simpleCacheAuth.csrftoken,
      ticket: simpleCacheAuth.ticket,
      nodeName: clusterInfo.data.now.name
    }),
  ])
  // Return
  let objectifyReturn = {
    identify: identify,
    host: host,
    success: true,
    needReAuth: false,
    node: clusterInfo.data,
    timeEnd: (new Date().getTime() - timeRequest)
  }
  if(!!Object.keys(rrddataInfo?.data)[0]) {
    objectifyReturn["rrddata"] = rrddataInfo?.data
  }
  if(!!Object.keys(statusInfo?.data)[0]) {
    objectifyReturn["status"] = statusInfo?.data
  }

  // Return
  return objectifyReturn
}

module.exports = {
  GetCrendetials,
  ClusterStatus,
  NodeStatus,
  ClusterRRDData,
  callAllRequest,
}