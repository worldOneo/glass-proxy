# Glass Proxy
## A simple but fast [TCP](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)-[Reverse-Proxy](https://en.wikipedia.org/wiki/Reverse_proxy) with following features:

 - Health Checking
 - Load Balancing

# Configuration
This is the default configuration (the file `glass.proxy.json`):
```json
{
    "addr": "0.0.0.0:25565",
    "hosts": [
        {
            "name": "Server-1",
            "addr": "localhost:25580"
        }
    ],
    "logConnections": true,
    "healthCheckSeconds": 5
}
```

| Value | Meaning |
| --- | --- |
| addr | The address to run the proxy on
| hosts | A list of hosts |
| (host) name | The name of the host  (for logging)
| (host) addr | The address of the host server
| logConnections | if the connections successful connections should be logged
| healthCheckSeconds | The time (in seconds) between server health checks

# Health Checks
The servers are checked regularly (based on the config `healthCheckSeconds`) if they can be reached. If not no client will be connected to that server.

# Load Balancing
The proxy selects a random server for every connection. This way the load will be (pseudo) randomly balanced between every registered host. The Health Checks ensure that the Server is reachable.