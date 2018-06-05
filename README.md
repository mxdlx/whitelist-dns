# Whitelist DNS
This is yet another DNS Server using [miekg's DNS Go library](https://github.com/miekg/dns). This is a really specific implementation to meet one requirement: a whitelist DNS Server. Also, as I wanted to get to know Redis, I used [go-redis](https://github.com/go-redis/redis). 

## Workflow
* Make a list with domains.
* Cache solved address.
* Push everything to Redis.
* Resolve with data in Redis.
* Think: this implementation must ignore so many RFCs.
