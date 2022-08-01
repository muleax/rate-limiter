# Overview
This project is for educational purposes only.

Simple HTTP rate-limiting service, based on a fixed window alghoritm: https://dev.to/satrobit/rate-limiting-using-the-fixed-window-algorithm-2hgm



    +----------------+            +----------------+            +----------------+
    |                |            |                |            |                |
    |     Client     |----------->|  Rate-limiter  |----------->|   App Service  |
    |                |            |                |            |                |
    +----------------+            +----------------+            +----------------+
                                          |
                                          |
                                          V
                                  +----------------+
                                  |                |
                                  |     Redis      |
                                  |                |
                                  +----------------+

Rate-limiter acts as a reverse proxy, and forward client requests to App service. Each request is expected to come with an API key, wich is used to identify the source.

For each incoming request Rate-limiter makes a Redis transaction of the form:

```
MULTI
    SET <api_key> 0 NX EX <timeout>
    INCR <api_key>
EXEC
```

Here timeout is equal to window size. The entry will be removed later, resetting the count.

Rate-limiter checks the request count for the key and decides whether to forward the request to App service.

# Performance

## 1. Simple setup

                                       +------------------------------------------------------+
    +---------------------+            |  +--------------+   +-------------+   +-----------+  |
    |                     |            |  | Rate-limiter |   | App Service |   |   Redis   |  |
    |       Vegeta        |----------->|  |   container  |   |  container  |   | container |  |
    |  (AWS m5zn.xlarge)  |            |  +--------------+   +-------------+   +-----------+  |
    +---------------------+            |                     (AWS m5zn.xlarge)                |
                                       +------------------------------------------------------+

- Window size: 5 seconds
- Limit per window per API key: 3 request
- 20 different API keys
- 50k connections

With such a setup, saturation point occurs around 55k RPS. Note that most requests failed with 429 status, but this is actually expected behavior.

```
ubuntu@ip-172-31-26-140:~$ cat ./attack.txt | vegeta attack -duration=15s -rate=100000 -max-workers=256 -connections=50000 | vegeta report
Requests      [total, rate, throughput]         826880, 55123.81, 12.00
Duration      [total, attack, wait]             15.003s, 15s, 2.831ms
Latencies     [min, mean, 50, 90, 95, 99, max]  657.952µs, 3.834ms, 3.597ms, 6.241ms, 7.291ms, 10.298ms, 24.725ms
Bytes In      [total, mean]                     15708020, 19.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           0.02%
Status Codes  [code:count]                      200:180  429:826700  
Error Set:
429 Too Many Requests
```

```
CONTAINER ID   NAME                          CPU %     MEM USAGE / LIMIT     MEM %
208981e821d7   rate-limiter-redis-1          88.47%    2.961MiB / 15.17GiB   0.02%
883d36e0e667   rate-limiter-rate-limiter-1   276.57%   24.32MiB / 15.17GiB   0.16%
0fbcba848de3   rate-limiter-app-server-1     0.01%     6.473MiB / 15.17GiB   0.04%
```

## 2. Clustered setup
EKS cluster set up on three m5zn.xlarge instances. Attack client runs on a separate machine.
                                                                                                                               
  +---------------------+       +------+       +---------------------+            +---------------+
  |                     |       |      |------>|                     |-------+--->|               |
  |       Vegeta        |------>|  LB  |       |    Rate-limiter     |       |    |  App Service  |
  |  (AWS m5zn.2xlarge) |       |      |---+   |  (AWS m5zn.xlarge)  |---+   |    |               |
  +---------------------+       +------+   |   +---------------------+   |   |    +---------------+
                                           |                             |   |                     
                                           |   +---------------------+   |   |                     
                                           |   |                     |-------+                     
                                           +-->|    Rate-limiter     |   |                        
                                               |  (AWS m5zn.xlarge)  |---+                        
                                               +---------------------+   |                        
                                                                         |                       
                                                                         V                        
                                                               +---------------------+
                                                               |                     |
                                                               |       Redis         |
                                                               |  (AWS m5zn.xlarge)  |
                                                               +---------------------+


CPU load is distributed evenly between pods, but for some reason, throughput doesn't exceed 60k RPS. Given that the Rate-limiter and Redis pods are underloaded, we must have hit some other bound.

```
[ec2-user@ip-172-31-24-45 ~]$ cat ./attack.txt | vegeta attack -duration=15s -rate=100000 -max-workers=256 -connections=50000 | vegeta report
Requests      [total, rate, throughput]         870656, 58044.00, 10.80
Duration      [total, attack, wait]             15.003s, 15s, 3.455ms
Latencies     [min, mean, 50, 90, 95, 99, max]  955.217µs, 4.213ms, 3.77ms, 6.626ms, 7.484ms, 9.448ms, 49.856ms
Bytes In      [total, mean]                     16540034, 19.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           0.02%
Status Codes  [code:count]                      200:162  429:870494  
Error Set:
429 Too Many Requests
```
```
muleax@MBP-Igor-2:~/rate-limiter$ kubectl top node
NAME                                              CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%   
ip-192-168-22-108.eu-central-1.compute.internal   1784m        45%    535Mi           3%        
ip-192-168-22-173.eu-central-1.compute.internal   723m         18%    573Mi           3%        
ip-192-168-42-137.eu-central-1.compute.internal   1848m        47%    534Mi           3% 
```
