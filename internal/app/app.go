package app

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/phogolabs/parcello"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/Nastradamus/list-ingress/internal/services/k8sservice"
)

const (
	templatePath = "templates/index.html"
)

type App struct {
	IngressService *k8sservice.IngressService
	KClientSet     *kubernetes.Clientset
	MainTemplate   *template.Template
}

func NewApp(ingressService *k8sservice.IngressService) *App {
	tmpl := getTemplate(templatePath)

	return &App{
		IngressService: ingressService,
		MainTemplate:   tmpl,
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
	//err = a.MainTemplate.Execute(w, vd)
	//if err != nil {
	//	klog.Exit(err)
	//}

	//w.WriteHeader(http.StatusOK)
}

func getTemplate(path string) *template.Template {
	file, err := parcello.Open(path)
	if err != nil {
		klog.Exit(err)
	}
	index, err := ioutil.ReadAll(file)
	if err != nil {
		klog.Exit(err)
	}

	tmpl, err := template.New("index").Parse(string(index))
	if err != nil {
		klog.Exit(err)
	}

	return tmpl
}
