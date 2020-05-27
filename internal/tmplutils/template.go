package tmplutils

import (
	"errors"
	"fmt"

	"github.com/Nastradamus/list-ingress/internal/core"
)

func Inc(val int) int {
	return val + 1
}

func MakeKubeDashLink(baseURL string) func(ingress core.IngressData) string {
	return func(ingress core.IngressData) string {
		return fmt.Sprintf("%v#/ingress/%v/%v?namespace=%v", baseURL, ingress.Namespace, ingress.Name, ingress.Namespace)
	}
}

func Dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid dict call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}
