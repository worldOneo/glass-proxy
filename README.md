# Glass Proxy
## A simple but fast UDP/TCP-[Reverse-Proxy](https://en.wikipedia.org/wiki/Reverse_proxy) with following features:

 - Health Checking
 - Load Balancing
 - Dynamically add/remove server

# Configuration
This is the default configuration (the file `glass.proxy.json`):
```json
{
    "protocol": "tcp",
    "addr": "0.0.0.0:25565",
    "interfaces": [],
    "hosts": [
        {
            "name": "Server-1",
            "addr": "localhost:25580"
        }
    ],
    "LogConfiguration": {
        "logConnections": true,
        "logDisconnect": false
    },
    "healthCheckSeconds": 5,
    "saveConfigOnClose": false
}
```

| Value | Meaning |
| --- | --- |
| protocol | The protocol the proxy should start. Currently supported: udp, udp4, udp6, tcp, tcp4, tcp6 |
| addr | The address to run the proxy on |
| interfaces | A list of network interfaces to use for out going connections. (If empty the default will be used) |
| hosts | A list of hosts |
| (host) name | The name of the host  (for logging)
| (host) addr | The address of the host server
| (LogConfiguration) logConnections | if the connections successful connections should be logged
| (LogConfiguration) logDisconnect | log when a connection is cloesed |
| healthCheckSeconds | The time (in seconds) between server health checks |
| UDPTimeout | The time (in ms) until a UDP connection is considered as closed |
# CLI
Some config-values can be set in the start command.
```
  -addr string
        The addr to start the server on. (default "0.0.0.0:25565")      
  -health int
        The time (in seconds) between health checks. (default 5)        
  -logc
        Log connections which where successfully bridged. (default true)
  -logd
        Log connections which where closed.
  -save
        Save the config when the server is stopped.
```
e.g: `$ ./glass-proxy -save=false -logc=true -health=3 -addr="0.0.0.0:1234"`  
(or IPv6): `$ ./glass-proxy -save=false -logc=true -health=3 -addr="[::]:1234"`


# Health Checks
The servers are checked regularly (based on the config `healthCheckSeconds`) if they can be reached (only one connection needed to verify). If not no client will be connected to that server.

# Load Balancing
The proxy selects a random server for every connection. This way the load will be (pseudo) randomly balanced between every registered host. The Health Checks ensure that the Server is reachable.

# Commands
While the proxy is running you can add/remove server
| cmd | Action |
| --- | --- |
| `add <Name> <addr>` | Add a server to the proxy which is then used in the Load Balancer |
| `rem <Name>` | Remove a server from the proxy (Opened connections will stay but no new connections will be created) |
| `list` | Lists all servers which are registered |
| `save` | Saves the config to the config file (Overwrites the old one)