(async () => {

    const host = "127.0.0.1"
    const port = 8007

    const agent_name = "test-agent"
    const rsa_key = "private.pem"
    const api_key = process.env.APIKEY
    const company_id = process.env.CUSTOMERID

    class SigningAgentClient {

        #agent_name = "test-agent"
        #rsa_key = null
        #api_key = null
        #rsa_key_pem = null
        #host = "localhost"
        #port = 8007
        #transaction_callback = async ()=>{}
        #company_id = null

        constructor(agent_name, rsa_key_file, api_key, company_id, host, port, transaction_callback) {
            const NodeRSA = require("node-rsa")
            const fs = require('fs')

            this.#agent_name = agent_name || this.#agent_name
            this.#api_key = api_key
            this.#company_id = company_id
            this.#host = host || this.#host
            this.#port = port || this.#port
            this.#transaction_callback = transaction_callback || this.#transaction_callback
            this.#rsa_key_pem = fs.readFileSync(rsa_key_file, "utf8")
            this.#rsa_key = NodeRSA(this.#rsa_key_pem, "pkcs1", { signingScheme: "pkcs1-sha256" })
        }

        async init() {
            let agent_id = await this.#registerAgent()
            this.#connectFeed()
            return agent_id
        }

        async #registerAgent() {
            // check if agent exists, if not create, if so return agentid
            let agent_id = await this.#executeAgentApiCall("GET", `http://${this.#host}:${this.#port}/api/v1/client`, null)
            if (agent_id.length > 0) {
                return agent_id[0]
            }

            // base64 rsa pem
            let b64pem = this.#base64Encode(this.#rsa_key_pem)

            // call register
            let result = await this.#executeAgentApiCall("POST", `http://${this.#host}:${this.#port}/api/v1/register`, {
                name: this.#agent_name,
                apikey: this.#api_key,
                base64privatekey: b64pem
            })

            // return agent_id
            if (result) {
                return result.agentId
            } else {
                return null
            }
        }

        #connectFeed() {
            let reconnect_timeout = null
            const WebSocket = require("ws").WebSocket
            const socket = new WebSocket(`ws://${this.#host}:${this.#port}/api/v1/client/feed`)

            let reconnect = (async function(){
                this.connectFeed()
                reconnect_timeout = null
            }).bind(this)
    
            socket.addEventListener("open", (event) => {
                console.log("feed connected")
            })
    
            socket.addEventListener("close", (event) => {
                console.log("feed disconnected")
                if (reconnect_timeout == null) {
                    reconnect_timeout = setTimeout(reconnect, 5000)
                }
            })
    
            socket.addEventListener("message", (event) => {
                
                if (typeof(this.#transaction_callback) == "function") {
                    let msg = JSON.parse(event.data)

                    // get transaction details
                    this.#getTransactionDetails(msg.id).then((details) => {
                        msg.details = details

                        // call callback
                        this.#transaction_callback(msg).then((approve) => {
                            if (approve) {
                                this.#approveTransaction(msg.id)
                            } else {
                                this.#rejectTransaction(msg.id)
                            }
                        })
                    })
                }
            })
        }

        #approveTransaction(transaction_id) {
            this.#executeAgentApiCall("PUT", `http://${this.#host}:${this.#port}/api/v1/client/action/${transaction_id}`, null)
        }

        #rejectTransaction(transaction_id) {
            this.#executeAgentApiCall("DELETE", `http://${this.#host}:${this.#port}/api/v1/client/action/${transaction_id}`, null)
        }

        #getTransactionDetails(msg) {
            if (!this.#company_id) {
                return null
            }

            let type_url = ""
            if (msg.type == "ApproveWithdraw") {
                type_url = "/withdraw"
            } else {
                type_url = "/transfer"
            }

            return this.#executePartnerApiCall("GET", `https://play-api.qredo.network/api/v1/p/company/${this.#company_id}/${type_url}/${msg.id}`, null)
        }

        async #executeAgentApiCall(method, url, body) {
            let json_body = undefined
            if (body != null) {
                json_body = JSON.stringify(body)
            }
    
            const req = {
                method: method, 
                mode: "cors", 
                cache: "no-cache", 
                credentials: "same-origin", 
                headers: {
                  "Content-Type": "application/json"                  
                },
                redirect: "follow", 
                referrerPolicy: "no-referrer", 
                body: json_body
            }

            let response = await fetch(url, req)
            if (response.status != 200) {
                console.error(response.status)
                return null
            }
    
            return await response.json()
        }

        async #executePartnerApiCall(method, url, body) {
            let json_body = null
            if (body != null) {
                json_body = JSON.stringify(body)
            }

            const req = {
                method: method, 
                mode: "cors", 
                cache: "no-cache", 
                credentials: "same-origin", 
                headers: {
                    "Content-Type": "application/json",
                    "x-api-key": this.#api_key
                },
                redirect: "follow", 
                referrerPolicy: "no-referrer", 
                body: json_body
            }

            const timestamp = Math.round(new Date().getTime() / 1000) + ""
            req.headers["x-timestamp"] = timestamp
            
            let to_sign = ""
            to_sign += timestamp
            to_sign += url
            if (json_body !== null) {
                to_sign += json_body
            }
            const rsa_sig = this.#rsa_key.sign(to_sign, 'buffer')
            const signature = this.#base64UrlEncode(signature)
            req.headers["x-sign"] = signature

            let response = fetch(url, req)
            if (response.status != 200) {
                return null
            }
    
            return await response.json()
        }

        #base64Encode(buffer){
            if (typeof(buffer) == "string") [
                buffer = Buffer.from(buffer, "utf-8")
            ]
    
            return btoa(String.fromCharCode(...new Uint8Array(buffer)))
        }

        #base64UrlEncode(buffer){
            this.#base64Encode(buffer)
                .replace(/\+/g, '-')
                .replace(/\//g, '_')
                .replace(/=+$/, '')
        }
    }

    const client = new SigningAgentClient(agent_name, rsa_key, api_key, company_id, host, port, async (trx) => {
        console.log(trx)

        console.log(details.statusDetails.netAmount)
        if (trx.details.statusDetails.netAmount < 100000) {
            // approve
            console.log("Approving transaction")
            return true
        }

        // ... more conditions

        // reject
        console.log("Rejecting transaction")
        return false
    })

    let agent_id = await client.init()
    console.log(agent_id)

})().then(()=>{})