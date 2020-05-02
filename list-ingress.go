package main

import (
	"flag"
	"net/http"
	"os"

	"k8s.io/klog"

	_ "github.com/Nastradamus/list-ingress/assets"
	"github.com/Nastradamus/list-ingress/internal/app"
	"github.com/Nastradamus/list-ingress/internal/services/k8sservice"
)

func main() {
	config := initConfig()
	printDiagMessage(config)

	klog.SetOutput(os.Stdout)

	ingressService := k8sservice.NewIngressService(config.K8sserviceConfig)

	app := app.NewApp(ingressService)

	http.HandleFunc("/", app.HandleRoot)

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
	flag.Parse()

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	for i, arg := range os.Args {
		if arg == klogFlags.Name() {
			err := klogFlags.Parse(os.Args[i+1:])
			if err != nil {
				klog.Exit(err)
			}
		}
	}

	return app.Config{
		K8sserviceConfig: k8sservice.Config{
			RunOutsideCluster: *runOutsideCluster,
		},
		LoggerFlags: *klogFlags,
	}
}

func printDiagMessage(c app.Config) {
	klog.V(0).Infof("Starting list-ingress...")
	klog.V(0).Infof("Verbosity level set to %v", c.LoggerFlags.Lookup("v").Value)
}
