package main

import (
	"flag"
	"fmt"
	ukube "github.com/Nastradamus/useless-operator/pkg/ukubernetes"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Ingress host
type Host struct {
	Name string
	Paths []string
}

// Ingress
type Ingress struct {
	Name string
	Namespace string
	Hosts []Host
}

func main() {
	v := flag.Int("v", 1, "Verbosity level (klog).")
	runOutsideCluster := flag.Bool("run-outside-cluster", false, "Set this flag when running " +
		"outside of the cluster.")

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	flag.Parse()
	klog.SetOutput(os.Stdout)

	// Get kubernetes config
	config, err := ukube.GetConfig(*runOutsideCluster)
	if err != nil {
		klog.Exit(err)
	}

	verbosity := klogFlags.Lookup("v")
	verbosity.Value.Set(strconv.Itoa(*v))

	klog.V(0).Infof("Starting list-ingress...")
	klog.V(0).Infof("Verbosity level set to %v", klogFlags.Lookup("v").Value)

	//Get tested k8s client
	kClient, err := ukube.GetKClient(config)
	if err != nil {
		klog.Exit(err)
	}

	// Serve HTTP
	finalHandler := http.HandlerFunc(handle)
	http.Handle("/", wrapper(finalHandler, kClient))

	klog.Infof("Starting HTTP server at http://0.0.0.0:8080")

	http.ListenAndServe(":8080", nil)

}

// getIngsStruct returns array of Ingress structures
func GetIngsStruct(ingInterface *v1beta1.IngressList) []Ingress {
	ingsStruct := make([]Ingress, len(ingInterface.Items))

	for ingNum, ingress := range ingInterface.Items {
		ns := ingress.Namespace
		name := ingress.Name

		ingsStruct[ingNum].Name = name
		ingsStruct[ingNum].Namespace = ns

		hostStruct := Host{} // think about pointer here

		for _, rule := range ingress.Spec.Rules {
			host := rule.Host

			hostStruct.Name = host
			if rule.IngressRuleValue.HTTP != nil {
				for _, pathStruct := range rule.IngressRuleValue.HTTP.Paths {
					//klog.V(0).Infof("Namespace: %q, Ingress: %q. Host: %q, Path: %q", ns, name, host, pathStruct.Path)
					hostStruct.Paths = append(hostStruct.Paths, pathStruct.Path)
				}
			}
		}
		ingsStruct[ingNum].Hosts = append(ingsStruct[ingNum].Hosts, hostStruct)
	}
	return ingsStruct
}

// We need wrapper (middleware) to pass arguments into http handler
// Main logic is here
func wrapper(h http.Handler, c *kubernetes.Clientset) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Server", "Anonymous Web Server 2.0")
		w.Header().Set("Content-Type", "text/html")


		// Part after / as a filter
		query := r.URL.Path[len("/"):]

		//fmt.Fprintf(w, "query: %v", query)
		if query == "" {
			fmt.Fprintf(w, "Find ingresses in k8s cluster. Please put search keyword as a GET query\n" +
				"(URL path)<br><br>Examples: \n<br> http://domain/my-ingress-name<br>\n" +
				"http://domain/ingress-domain <br>\n" +
				"http://domain/Ingress  (will show all ingresses in the cluster) <br><br>\n\n")
		}

		// Get all ingresses
		ingInterface, err := c.ExtensionsV1beta1().Ingresses("").List(metav1.ListOptions{})
		if err != nil {
			w.WriteHeader(500)
		}

		ingresses := GetIngsStruct(ingInterface)

		line := ""
		for _, ingress := range ingresses {
			for _, host := range ingress.Hosts {
				for _, path := range host.Paths {
					line += `"Namespace: ` + ingress.Namespace + ", Ingress: " + ingress.Name +
						", Domain: " + host.Name + ", Path: " + path + `"` + "<br>\n"
				}
			}
		}

		// Full text search :-)
		if query != "" {
			fmt.Fprintf(w, "<br>Query: %q <br><br>", query)
			lines := strings.Split(line, "\n")
			for _, curLine := range lines {
				if strings.Contains(curLine, query) {
					fmt.Fprint(w, curLine)
				}
			}
		}

		h.ServeHTTP(w, r)
	})
}
// We need empty function to pass arguments into http.Handle()
func handle(w http.ResponseWriter, r *http.Request) {
	return
}