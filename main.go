package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/xaxes/transfigurator/transform"
)

func transformMany(path string, transformers []transform.Transformer) ([]string, error) {
	var prURLs []string
	for _, trans := range transformers {
		url, err := trans.Apply(path)
		if err != nil {
			return nil, errors.Wrap(err, trans.Name())
		}
		prURLs = append(prURLs, url)
	}

	return prURLs, nil
}

func applyAll(wg *sync.WaitGroup, gh *github.Client, ghToken, org string, repo repository) {
	transSet := transform.GetAll(gh, ghToken, org)
	var names []string
	for _, trans := range transSet {
		names = append(names, trans.Name())
	}
	// log.Printf("Applying %v on %s", strings.Join(names, ", "), *repo.Name)
	urls, err := transformMany(repo.Path(), transSet)
	if err != nil {
		log.Printf("Failed to apply transformer on %s: %s", repo.Path(), err)
	}
	log.Printf("Pull requests for %s: %s", *repo.Name, strings.Join(urls, ", "))

	wg.Done()
}

func apply(wg *sync.WaitGroup, repo repository, ts ...transform.Transformer) {
	var names []string
	for _, t := range ts {
		names = append(names, t.Name())
	}
	log.Printf("Applying %v on %s", strings.Join(names, ", "), *repo.Name)
	urls, err := transformMany(repo.Path(), ts)
	if err != nil {
		log.Printf("Failed to apply transformer on %s: %s", repo.Path(), err)
	}
	log.Printf("Pull requests for %s: %v", *repo.Name, urls)

	wg.Done()
}

type transList []string

func (t *transList) Set(v string) error {
	*t = strings.Split(v, ",")

	return nil
}

func (t transList) String() string {
	return strings.Join(t, ",")
}

func renderConfig() *Config {
	var transformers transList
	flag.Var(&transformers, "", "Comma-separated list of transformers")
	org := flag.String("org", "", "Github organization")
	token := flag.String("token", "", "Github access token")
	configFile := flag.String("config", "config.yml", "Configuration file")
	flag.Parse()

	f, err := ioutil.ReadFile(*configFile)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("open %s: %s", *configFile, err)
	}

	var c Config
	if err := yaml.Unmarshal(f, &c); err != nil {
		log.Fatalf("unmarshal %s: %s", *configFile, err)
	}

	if *org != "" {
		c.Github.Organization = *org
	}
	if *token != "" {
		c.Github.AccessToken = *token
	}
	if len(transformers) != 0 {
		c.Transformers = transformers
	}

	return &c
}

func main() {
	c := renderConfig()

	ctx := context.Background()
	client := newClient(ctx, c.Github.AccessToken)
	repos, err := gatherOrgRepos(ctx, client, c.Github.Organization)
	if err != nil {
		log.Fatal(err)
	}

	cloned, err := getRepos(repos, c.LocalStorage)
	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}
	for _, repo := range cloned {
		wg.Add(1)
		go apply(&wg, repo, transform.NewHCLFMT(client, c.Github.AccessToken, c.Github.Organization))
	}
	wg.Wait()
}
