#### UDP Proxy + Load Balancer

**Why required -** 

Swiftwave uses HAProxy as ingress and it can handle TCP/HTTP traffic. But it can't handle UDP traffic. So, we need a UDP proxy to handle UDP traffic and also to load balance it.

You may thought why we not using Docker Swarm Ingress for UDP traffic.

The reason is
- we need some control over the port mapping.
- we need something which can monitor traffic and can take action based on that. (future requirement)
- we have some feature in our roadmap which required UDP proxy. (future requirement)

**How it works -**
- It will listen on a port and forward the traffic to the destination based on the configuration.
- It will have a backup mechanism to alert swiftwave if service is unreachable. (future requirement)
- Will support runtime configuration change.
- Will only managed over unix socket.

**Environment Variables -**
- `SOCKET_PATH` - Path of the unix socket. Default: `/etc/udplb/api.sock`
- `RECORDS_PATH` - Path of the records file. Default: `/etc/udplb/records`
