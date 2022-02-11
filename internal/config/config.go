package config

type Config struct {
	HelmVersion      string
	KustomizeVersion string
	ArgoCDToken      string
	GitLabToken      string
	RepoURL          string
	MasterBranch     string
	Branch           string
	Commit           string
}

func Load() *Config {
	cfg := &Config{
		HelmVersion:      "2.16.12",
		KustomizeVersion: "4.5.2",
		ArgoCDToken:      "",
		GitLabToken:      "",
		RepoURL:          "",
		MasterBranch:     "",
		Branch:           "",
		Commit:           "",
	}
	return cfg
}
