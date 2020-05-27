package app

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/phogolabs/parcello"
	"k8s.io/klog"

	"github.com/Nastradamus/list-ingress/internal/services/k8sservice"
	"github.com/Nastradamus/list-ingress/internal/tmplutils"
)

const (
	templatePath = "templates/index.html"
)

type App struct {
	IngressService *k8sservice.IngressService
	MainTemplate   *template.Template
	Config         Config
}

func NewApp(config Config) *App {
	tmpl := getTemplate(templatePath, config)
	ingressService := k8sservice.NewIngressService(config.K8sserviceConfig)

	return &App{
		IngressService: ingressService,
		MainTemplate:   tmpl,
		Config:         config,
	}
}

func (a *App) HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Anonymous Web Server 2.0")
	w.Header().Set("Content-Type", "text/html")

	query := strings.TrimPrefix(r.URL.Path, "/")

	ingresses, intersections, err := a.IngressService.GetFilteredIngressesWIntersections(query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		klog.Exit(err)
	}

	vd := ViewData{
		Search:                query,
		KubeDashURL:           a.Config.KubeDashURL,
		Ingresses:             ingresses,
		IntersectionIngresses: intersections,
	}

	var buf bytes.Buffer
	if err := a.MainTemplate.Execute(&buf, vd); err != nil {
		klog.Exit(err)
	} else {
		_, err := io.Copy(w, &buf)
		if err != nil {
			klog.Exit(err)
		}
	}
}

func (a App) HandleHealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func getTemplate(path string, config Config) *template.Template {
	file, err := parcello.Open(path)
	if err != nil {
		klog.Exit(err)
	}
	index, err := ioutil.ReadAll(file)
	if err != nil {
		klog.Exit(err)
	}

	fns := template.FuncMap{
		"inc":          tmplutils.Inc,
		"kubeDashLink": tmplutils.MakeKubeDashLink(config.KubeDashURL),
		"dict":         tmplutils.Dict,
	}
	tmpl := template.Must(template.New("index").Funcs(fns).Parse(string(index)))

	return tmpl
}
