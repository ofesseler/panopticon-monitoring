# Panopticon-Monitoring

Panopticon-Monitoring ist ein Tool, dass die Ausgabe von Prometheus für Docker, Docker Swarm, GlusterFS, Consul und Weave Net des Clusters bewertet und einen aktuellen Status darstellt.

## Installation
Um Panopticon-Monitoring betreiben zu können, ist eine Prometheus Instanz notwendig, die die Exporter für Docker, Docker Swarm, GlusterFS, Consul und Weave Net abruft.

```
go get github.com/ofesseler/panopticon-monitoring
```


| Flag | Bemerkung |
|------|-----------|
| cluster-nodes | Anzahl der im Cluster befindlichen Knoten.|
| prom-host | Hostname und Port der Prometheus Instanz |
| listen-address | Port und Adresse unter der Panopticon-Monitoring erreichbar sein soll. |

## Tests

Es gibt die Möglichkeit, das Tool zu testen. Mit `make docker` im Projektroot, startet ein Docker Container mit Prometheus. Die von Proemtheus abgefragten Metrikfiles werden aus dem Projektverzeichnis `docker-assets/metrics` abgerufen und sind im Betrieb veränderbar. Dadurch lässt sich das Verhalten im Fehlerfall simulieren.

## Hintergrund

Das Tool ist in Verbindung mit meiner Bachelorarbeit entstanden, die unter https://github.com/ofesseler/monitoring-docker-swarm-thesis/ ab dem 1.4.2017 eingesehen werden kann. 
