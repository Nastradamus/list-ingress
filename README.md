# list-ingress
A simple http application which helps to find all Ingresses in the cluster (helps you find intersections)

### Example of usage:
```bash
go build && ./list-ingress -run-outside-cluster

I0203 15:09:36.514766    5978 list-ingress.go:49] Starting list-ingress...
I0203 15:09:36.515393    5978 list-ingress.go:50] Verbosity level set to 1
I0203 15:09:37.411819    5978 kubernetes.go:53] There are 43 nodes in the cluster
I0203 15:09:37.411839    5978 list-ingress.go:62] Starting HTTP server at http://0.0.0.0:8080
```