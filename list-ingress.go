package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	ukube "github.com/Nastradamus/useless-operator/pkg/ukubernetes"
)

// Ingress rule
type Rule struct {
	Host  string
	Paths []string
}

// Ingress
type Ingress struct {
	Name      string
	Namespace string
	Rules     []Rule
}

func main() {
	v := flag.Int("v", 1, "Verbosity level (klog).")
	runOutsideCluster := flag.Bool("run-outside-cluster", false, "Set this flag when running "+
		"outside of the cluster.")

	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)

	flag.Parse()

	klog.InitFlags(klogFlags)
	klog.SetOutput(os.Stdout)

	verbosity := klogFlags.Lookup("v")
	verbosity.Value.Set(strconv.Itoa(*v))

	// Get kubernetes config
	config, err := ukube.GetConfig(*runOutsideCluster)
	if err != nil {
		klog.Exit(err)
	}

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

		for _, rule := range ingress.Spec.Rules {

			ruleStruct := Rule{}
			ruleStruct.Host = rule.Host

			if rule.IngressRuleValue.HTTP != nil {
				for _, pathStruct := range rule.IngressRuleValue.HTTP.Paths {
					//klog.V(0).Infof("Namespace: %q, Ingress: %q. Rule: %q, Path: %q", ns, name, host, pathStruct.Path)
					ruleStruct.Paths = append(ruleStruct.Paths, pathStruct.Path)
				}
			}

			ingsStruct[ingNum].Rules = append(ingsStruct[ingNum].Rules, ruleStruct)
		}
	}
	return ingsStruct
}

// We need wrapper (middleware) to pass arguments into http handler
// Main logic is here
func wrapper(h http.Handler, c *kubernetes.Clientset) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Server", "Anonymous Web Server 2.0")
		w.Header().Set("Content-Type", "text/html")

		// Get all ingresses into IngressList struct
		myIngressesList, err := c.ExtensionsV1beta1().Ingresses("").List(metav1.ListOptions{})
		if err != nil {
			w.WriteHeader(500)
		}
		ingresses := GetIngsStruct(myIngressesList)

		// Part after / as a filter
		query := r.URL.Path[len("/"):]

		var ingInterCnt int // Ingress intersections count

		if query == "" {
			fmt.Fprintf(w, "Find ingresses in k8s cluster. Please put search keyword as a GET query\n"+
				"(URL path)<br><br>Examples: \n<br> http://domain/my-ingress-name<br>\n"+
				"http://domain/ingress-domain <br>\n"+
				"http://domain/Ingress  (will show all ingresses in the cluster) <br><br>\n\n")

			fmt.Fprintf(w, "<b><font color=\"red\">Ingress intersections</font></b>:<br><br>\n")

			// Find ingress intersections (linear)
			for _, ingOuter := range ingresses {
				for _, ingInner := range ingresses {
					if ingOuter.Name == ingInner.Name {
						break
					}

					if areIngressesIntersects(ingOuter, ingInner) {
						ingInterCnt += 1

						lineA := "" // ingOuter
						lineB := "" // ingInner
						for _, host := range ingOuter.Rules {
							for _, path := range host.Paths {
								lineA += `"Namespace: ` + ingOuter.Namespace + ", Ingress: " + ingOuter.Name +
									", Domain: " + host.Host + ", Path: " + path + `"` + "<br>\n"
							}
						}
						for _, host := range ingInner.Rules {
							for _, path := range host.Paths {
								lineB += `"Namespace: ` + ingInner.Namespace + ", Ingress: " + ingInner.Name +
									", Domain: " + host.Host + ", Path: " + path + `"` + "<br>\n"
							}
						}

						fmt.Fprintf(w, "Ingress A:<br>\n")
						fmt.Fprintf(w, "%v\n", lineA)
						fmt.Fprintf(w, "Ingress B:<br>\n")
						fmt.Fprintf(w, "%v<br>\n", lineB)
					}
				}
			}
			fmt.Fprintf(w, "\n\n<br<br> Intersections count: <font color=\"red\">%d</font><br>\n", ingInterCnt)
		}

		// If query is non-empty, make search

		// First, get all ingresses slitted into lines
		line := ""
		for _, ingress := range ingresses {
			for _, host := range ingress.Rules {
				for _, path := range host.Paths {
					line += "<tr>"
					line += "<td>"
					line += ingress.Namespace
					line += "</td>"

					line += "<td>"
					line += ingress.Name
					line += "</td>"

					line += "<td>"
					line += host.Host
					line += "</td>"

					line += "<td>"
					line += path
					line += "</td>"

					line += "</tr>\n"
				}
			}
		}

		// Full text search :-)
		if query != "" {
			fmt.Fprintf(w, "<br>Query: %q <br><br>", query)

			fmt.Fprintf(w, "<table>")

			writeTableHead(&w)

			lines := strings.Split(line, "\n")
			for _, curLine := range lines {
				if strings.Contains(curLine, query) {
					fmt.Fprint(w, curLine)
				}
			}
			fmt.Fprintf(w, "</table>")
		}

		h.ServeHTTP(w, r)
	})
}

func writeTableHead(w *http.ResponseWriter) {
	fmt.Fprintf(*w, "<head>")
	fmt.Fprintf(*w, "<tr>")

	fmt.Fprintf(*w, "<td>")
	fmt.Fprintf(*w, "Namespace")
	fmt.Fprintf(*w, "</td>")

	fmt.Fprintf(*w, "<td>")
	fmt.Fprintf(*w, "Ingress")
	fmt.Fprintf(*w, "</td>")

	fmt.Fprintf(*w, "<td>")
	fmt.Fprintf(*w, "Domain")
	fmt.Fprintf(*w, "</td>")

	fmt.Fprintf(*w, "<td>")
	fmt.Fprintf(*w, "Path")
	fmt.Fprintf(*w, "</td>")

	fmt.Fprintf(*w, "</tr>")
	fmt.Fprintf(*w, "</head>")
}

// We need empty function to pass arguments into http.Handle()
func handle(w http.ResponseWriter, r *http.Request) {
	return
}

// areIngressesIntersects compares two Ingresses for host + path intersections
func areIngressesIntersects(a, b Ingress) bool {

	hostIntersects := false
	pathIntersects := false

	// Find Ingress host intersection
	for _, hostA := range a.Rules {
		for _, hostB := range b.Rules {

			// Compare hostnames
			if hostA.Host == hostB.Host {
				hostIntersects = true

				// If host intersects, let's check paths
				for _, pathA := range hostA.Paths {
					for _, pathB := range hostB.Paths {
						if pathA == pathB {
							pathIntersects = true
							break
						}
					}
					if pathIntersects {
						break
					}
				}

				// Found
				if hostIntersects && pathIntersects {
					return true
				}

			}
		}
	}

	return false
}
