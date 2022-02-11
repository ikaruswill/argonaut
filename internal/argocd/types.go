package argocd

type Metadata struct {
	Name string `yaml:"name"`
}

type Helm struct {
	ReleaseName string `yaml:"releaseName,omitempty"`
	Version     string `yaml:"version"`
	Values      string `yaml:"values,omitempty"`
}

type Source struct {
	Path string `yaml:"path,omitempty"`
	Helm *Helm  `yaml:"helm,omitempty"`
}

type Spec struct {
	Source *Source `yaml:"source"`
}

type Application struct {
	ApiVersion string    `yaml:"apiVersion"`
	Kind       string    `yaml:"kind"`
	Metadata   *Metadata `yaml:"metadata"`
	Spec       *Spec     `yaml:"spec"`
}
