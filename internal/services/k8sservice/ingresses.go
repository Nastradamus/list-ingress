package k8sservice

import (
	"sort"
	"strings"

	ukube "github.com/Nastradamus/useless-operator/pkg/ukubernetes"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	"github.com/Nastradamus/list-ingress/internal/core"
)

type Config struct {
	RunOutsideCluster bool
}

func NewKClient(c Config) *kubernetes.Clientset {
	// Get kubernetes config
	uKubeConfig, err := ukube.GetConfig(c.RunOutsideCluster)
	if err != nil {
		klog.Exit(err)
	}

	//Get tested k8s client
	kClient, err := ukube.GetKClient(uKubeConfig)
	if err != nil {
		klog.Exit(err)
	}

	return kClient
}

func NewIngressService(c Config) *IngressService {
	return &IngressService{
		KClientSet: NewKClient(c),
	}
}

type IngressService struct {
	KClientSet *kubernetes.Clientset
}

func (is IngressService) GetIngresses() ([]core.IngressData, error) {
	myIngressesList, err := is.KClientSet.ExtensionsV1beta1().Ingresses("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	ingresses := getIngressData(myIngressesList)

	sort.Slice(ingresses, makeIngressSorter(ingresses))

	return ingresses, nil
}

func (is IngressService) GetFilteredIngresses(filter string) ([]core.IngressData, error) {
	ingresses, err := is.GetIngresses()
	if err != nil {
		return nil, err
	}

	return filterIngresses(ingresses, filter), nil
}

func (is IngressService) GetFilteredIngressesWIntersections(filter string) ([]core.IngressData, []core.IngressData, error) {
	ingresses, err := is.GetFilteredIngresses(filter)
	if err != nil {
		return nil, nil, err
	}

	intersections := findIntersections(ingresses)

	return ingresses, intersections, nil
}

// Input ingresses must be sorted!
func findIntersections(ingresses []core.IngressData) []core.IngressData {
	intersections := make([]core.IngressData, 0)
	prevAdded := false

	for i, ingress := range ingresses {
		if i == 0 {
			continue
		}

		if isIntersect(ingresses[i-1], ingress) {
			if !prevAdded {
				intersections = append(intersections, ingresses[i-1])

				prevAdded = true
			}
			intersections = append(intersections, ingress)

			continue
		}
		prevAdded = false
	}

	return intersections
}

func getIngressData(ingInterface *v1beta1.IngressList) []core.IngressData {
	ingresses := make([]core.IngressData, 0, len(ingInterface.Items))

	for _, ingress := range ingInterface.Items {
		for _, rule := range ingress.Spec.Rules {
			if rule.IngressRuleValue.HTTP != nil {
				for _, pathStruct := range rule.IngressRuleValue.HTTP.Paths {
					ingress := core.IngressData{
						Name:          ingress.Name,
						Namespace:     ingress.Namespace,
						Host:          rule.Host,
						Path:          pathStruct.Path,
						RewriteTarget: ingress.Annotations["nginx.ingress.kubernetes.io/rewrite-target"],
					}

					ingresses = append(ingresses, ingress)
				}
			} else {
				ingress := core.IngressData{
					Name:      ingress.Name,
					Namespace: ingress.Namespace,
					Host:      rule.Host,
				}
				ingresses = append(ingresses, ingress)
			}
		}
	}

	return ingresses
}

func filterIngresses(ingresses []core.IngressData, query string) []core.IngressData {
	filtered := ingresses[:0]
	for _, ingress := range ingresses {
		if ingressFilterByOccurrence(ingress, query) {
			filtered = append(filtered, ingress)
		}
	}

	return filtered
}

func ingressFilterByOccurrence(i core.IngressData, filter string) bool {
	if strings.Contains(i.Namespace, filter) ||
		strings.Contains(i.Name, filter) ||
		strings.Contains(i.Host, filter) ||
		strings.Contains(i.Path, filter) ||
		strings.Contains(i.RewriteTarget, filter) {
		return true
	}

	return false
}

func makeIngressSorter(ingresses []core.IngressData) func(i, j int) bool {
	return func(i, j int) bool {
		byPath := ingresses[i].Namespace == ingresses[j].Namespace && ingresses[i].Name == ingresses[j].Name && ingresses[i].Host == ingresses[j].Host
		byHost := ingresses[i].Namespace == ingresses[j].Namespace && ingresses[i].Name == ingresses[j].Name
		byName := ingresses[i].Namespace == ingresses[j].Namespace

		switch {
		case byPath:
			return ingresses[i].Path < ingresses[j].Path
		case byHost:
			return ingresses[i].Host < ingresses[j].Host
		case byName:
			return ingresses[i].Name < ingresses[j].Name
		default:
			return ingresses[i].Namespace < ingresses[j].Namespace
		}
	}
}

func isIntersect(ingressA, ingressB core.IngressData) bool {
	if ingressA.Host == ingressB.Host && ingressA.Path == ingressB.Path {
		return true
	}

	return false
}
