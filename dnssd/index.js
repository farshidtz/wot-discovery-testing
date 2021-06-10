const MDNS = require('multicast-dns');
const fetch = require('node-fetch');
class MDNSWoTTestSuite {
    constructor() {
        this.mdns = MDNS();
    }

    testDiscovererThing(timeout) {
        return new Promise((resolve, reject) => {
            const timer = setTimeout(() => {
                this.mdns.removeListener('query', handler)
                reject(new Error("timeout"))
            }, timeout);

            function handler(query) {
                for (const question of query.questions) {
                    if (question.name.startsWith("_wot")) {
                        clearTimeout(timer)
                        this.removeListener('query', handler)
                        resolve()
                    }
                }
            }
            this.mdns.on('query', handler)
        })
    }

    testDiscovererDirectory(timeout) {
        return new Promise((resolve, reject) => {
            const timer = setTimeout(() => {
                this.mdns.removeListener('query', handler)
                reject(new Error("timeout"))
            }, timeout);

            function handler(query) {
                for (const question of query.questions) {
                    if (question.name.startsWith("_directory._sub._wot")) {
                        clearTimeout(timer)
                        this.removeListener('query', handler)
                        resolve(true)
                    }
                }
            }
            this.mdns.on('query', handler)
        })
    }

    testDiscovereeThing(timeout) {
        return new Promise((resolve, reject) => {


            const interval = setInterval(() => {
                this.mdns.query('_wot._tcp.local', 'PTR')
            }, 1000);
            async function handler(packet) {
                const ans = packet.answers.find(ans => ans.type === 'PTR' && ans.name === '_wot._tcp.local')
                if (ans) {
                    const svr = packet.additionals.find(ans => ans.type === 'SRV')
                    const txt = packet.additionals.find(ans => ans.type === 'TXT')
                    const port = svr && svr.data.port
                    const host = svr && svr.data.target

                    const rawTxtData = txt && txt.data.toString()
                    const regex = /^td=(?<path>.*),type=(?<type>.*)/;
                    const match = regex.exec(rawTxtData)
                    if (match) {
                        clearTimeout(timer)
                        clearInterval(interval)
                        this.removeListener('response', handler)

                        const path = match.groups.path
                        const type = match.groups.type

                        if (type !== 'Thing') {
                            reject(new Error("Wrong type advertised:", type))
                            return
                        }

                        const fullPath = `http://${host}:${port}${path}`

                        const res = await fetch(fullPath)

                        if (res.status !== 200) {
                            reject(new Error("Wrong address advertised:", fullPath))
                            return
                        }

                        resolve()
                    }
                }
            }

            this.mdns.on('response', handler)
            const timer = setTimeout(() => {
                clearInterval(interval)
                this.mdns.removeListener('response', handler);
                reject(new Error("timeout"))
            }, timeout);
        })
    }

    testDiscovereeDirectory(timeout) {
        return new Promise((resolve, reject) => {
            const interval = setInterval(() => {
                this.mdns.query('_directory._sub._wot', 'PTR')
            }, 1000);

            async function handler(packet) {
                const ans = packet.answers.find(ans => ans.type === 'PTR' && ans.name === '_directory._sub._wot')
                if (ans) {
                    const svr = packet.additionals.find(ans => ans.type === 'SRV')
                    const txt = packet.additionals.find(ans => ans.type === 'TXT')
                    const port = svr && svr.data.port
                    const host = svr && svr.data.target

                    const rawTxtData = txt && txt.data.toString()
                    const regex = /^td=(?<path>.*),type=(?<type>.*)/;
                    const match = regex.exec(rawTxtData)
                    if (match) {
                        clearTimeout(timer)
                        clearInterval(interval)
                        this.removeListener('response', handler)

                        const path = match.groups.path
                        const type = match.groups.type

                        if (type !== 'Directory') {
                            reject(new Error("Wrong type advertised:", type))
                            return
                        }

                        const fullPath = `http://${host}:${port}${path}`

                        const res = await fetch(fullPath)

                        if (res.status !== 200) {
                            reject(new Error("Wrong address advertised:", fullPath))
                            return
                        }

                        resolve()
                    }
                }
            }

            this.mdns.on('response', handler)
            const timer = setTimeout(() => {
                clearInterval(interval)
                this.mdns.removeListener('response', handler);
                reject(new Error("timeout"))
            }, timeout);
        })
    }

    close(callback) {
        this.mdns.removeAllListeners()
        this.mdns.destroy(callback)
    }
}


module.exports = MDNSWoTTestSuite