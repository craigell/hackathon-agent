# 🎉🎉🎉 Mock Management OTel Collector -- Hackathon Edition🎉🎉🎉

There are 2 images that need to be built in order to use the mock management OTel collector
* Agent with NGINX Plus image
* Custom OTel collector image

To build & run these images run the following
```
> make run-mock-collector-hackathon-edition
```

Once everything is started there should be 5 containers running
```
CONTAINER ID  IMAGE                                                   COMMAND               CREATED         STATUS         PORTS                                         
f4e7264ceb0b  localhost/nginx_plus_agent_ubuntu_22.04:latest                                15 seconds ago  Up 16 seconds  80/tcp, 443/tcp                                 mock-collector-agent-with-nginx-plus
04a3bce14c7a  localhost/mock-collector:latest                         /mock-management-...  12 seconds ago  Up 13 seconds  0.0.0.0:4320->4317/tcp, 0.0.0.0:9775->9090/tcp  mock-collector-otel-collector
8efe150cbd96  docker.io/prom/prometheus:latest                        --config.file=/et...  11 seconds ago  Up 11 seconds  0.0.0.0:9090->9090/tcp, 9090/tcp                mock-collector-prometheus
e55572131d2a  docker.io/grafana/grafana:latest                                              9 seconds ago   Up 10 seconds  0.0.0.0:3002->3000/tcp, 3000/tcp                mock-collector-grafana
69c8efab591b  docker.io/grafana/loki:latest                           -config.file=/etc...  8 seconds ago   Up 8 seconds   0.0.0.0:3100->3100/tcp, 3100/tcp                mock-collector-loki

```

To view the metrics & logs, the grafana UI can be used by accessing this URL http://localhost:3002/login (Note: username/password is admin/admin)

To stop everything run the following
```
make stop-mock-management-otel-collector-hackathon-edition
```
