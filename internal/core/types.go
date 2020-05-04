package core

type IngressData struct {
	Name          string
	Namespace     string
	Host          string
	Path          string
	RewriteTarget string
}
