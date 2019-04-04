package transform

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/google/go-github/github"
)

type NomadDistinct struct {
	gh      *github.Client
	ghToken string
	org     string
}

func (t NomadDistinct) Name() string {
	return "Nomad Distinct"
}

func (t NomadDistinct) Desc() string {
	return "Remove nomad distinct constraints"
}

func isNomadFile(path string) bool {
	if filepath.Ext(path) == ".nomad" || filepath.Ext(path) == ".hcl" {
		return true
	}
	return false
}

var (
	constraint = regexp.MustCompile(`(?m)(^[ \t]*constraint\s*{\s*$
		^\s*operator\s*=\s*\"distinct_[a-zA-Z]+\"\s*$
		(^\s*.*\s*$)*
		^\s*}[ \t]*$)`)
)

func (t NomadDistinct) removeConstraint(path string) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	replaced := constraint.ReplaceAllString(string(content), "")
	if replaced != string(content) {
		log.Printf("Removed constraint from %s", path)
	}

	return nil
	// return ioutil.WriteFile(path, []byte(replaced), 0600)
}

func (t NomadDistinct) Apply(path string) (string, error) {
	var files []string
	if err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error on %s: %s", file, err)
			return nil
		}
		if !skip(file, info) && isNomadFile(file) {
			log.Printf("%s to be patched", file)
			files = append(files, file)
		}
		return nil
	}); err != nil {
		return "", err
	}

	for _, f := range files {
		if err := t.removeConstraint(f); err != nil {
			log.Printf("Fail on %s: %s", f, err)
			return "", err
		}
	}

	return handleGit(t.gh, t.ghToken, t.org, path, t.Desc())
}

func NewNomadDistinct(gh *github.Client, ghToken, org string) NomadDistinct {
	return NomadDistinct{
		gh:      gh,
		ghToken: ghToken,
		org:     org,
	}
}
