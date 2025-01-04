# atest-ext-collector
Collector Extension of [API Testing](https://github.com/LinuxSuRen/api-testing)

## HTTP Proxy

Below is the command to start the HTTP proxy server.

```shell
atest-collector proxy
```
## DNS Server

```shell
atest-collector dns --simple-config config.yaml
```

Below is an example of a simple DNS config:
```yaml
simple:
    atest.com: 127.0.0.1
    www.atest.com: 127.0.0.1
```
