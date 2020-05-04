package app

import (
	"flag"

	"github.com/Nastradamus/list-ingress/internal/core"
	"github.com/Nastradamus/list-ingress/internal/services/k8sservice"
)

type Config struct {
	K8sserviceConfig k8sservice.Config
	LoggerFlags      flag.FlagSet
}

type ViewData struct {
	Search                string
	Ingresses             []core.IngressData
	IntersectionIngresses []core.IngressData
}
