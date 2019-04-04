package main

type Github struct {
	AccessToken  string `yaml:"access-token"`
	Organization string
}

type Git struct {
	Author string
	Email  string
}

type Config struct {
	Github       Github
	Git          Git
	LocalStorage string `yaml:"local-storage"`
	Transformers []string
}
