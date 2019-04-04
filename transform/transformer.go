package transform

import (
	"log"
	"os"
	"regexp"

	"github.com/google/go-github/github"
)

type Transformer interface {
	Name() string
	Desc() string
	Apply(string) (string, error)
}

func GetAll(gh *github.Client, ghToken, org string) []Transformer {
	return []Transformer{
		NewNomadDistinct(gh, ghToken, org),
		NewHCLFMT(gh, ghToken, org),
	}
}

var (
	initialised bool
	checks      []*regexp.Regexp
)

var regexps = []string{
	`.*\.git.*`,
	`.*/vendor/.*`,
}

func initChecks() {
	if initialised {
		return
	}

	for _, r := range regexps {
		compiled, err := regexp.Compile(r)
		if err != nil {
			log.Fatalf(`regexp "%s" compilation failed: %s`, r, err)
		}
		checks = append(checks, compiled)
	}

	initialised = true
}

func skip(path string, info os.FileInfo) bool {
	if info.IsDir() {
		// log.Printf("%s is dir", path)
		return true
	}

	initChecks()
	for _, r := range checks {
		if r.MatchString(path) {
			// log.Printf("%s matched regexp", path)
			return true
		}
	}

	// log.Printf("%s not skipped", path)
	return false
}
