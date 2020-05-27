package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	"k8s.io/klog"

	_ "github.com/Nastradamus/list-ingress/assets"
	"github.com/Nastradamus/list-ingress/internal/app"
	"github.com/Nastradamus/list-ingress/internal/services/k8sservice"
)

func main() {
	config := initConfig()
	printDiagMessage(config)

	app := app.NewApp(config)

	http.HandleFunc("/", app.HandleRoot)
	http.HandleFunc("/healthz", app.HandleHealthCheck)

	klog.Infof("Starting HTTP server at http://0.0.0.0:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		klog.Exit(err)
	}
}

func initConfig() app.Config {
	runOutsideCluster := flag.Bool(
		"run-outside-cluster",
		false,
		"Set this flag when running outside of the cluster.",
	)

	kubeDashURL := flag.String(
		"kube-dash-url",
		"",
		"Base URL of kubernetes dashboard",
	)

	verbosity := flag.Int("v", 1, "Verbosity level (klog).")
	flag.Parse()

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	klog.SetOutput(os.Stdout)
	err := klogFlags.Lookup("v").Value.Set(strconv.Itoa(*verbosity))
	if err != nil {
		klog.Exit(err)
	}

	return app.Config{
		K8sserviceConfig: k8sservice.Config{
			RunOutsideCluster: *runOutsideCluster,
		},
		LoggerFlags: *klogFlags,
		KubeDashURL: *kubeDashURL,
	}
}

func printDiagMessage(c app.Config) {
	klog.V(0).Infof("Starting list-ingress...")
	klog.V(0).Infof("Verbosity level set to %v", c.LoggerFlags.Lookup("v").Value)
	klog.V(0).Infof("Kubernetes dashboard base url: %v", c.KubeDashURL)
}
