package transform

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
)

type HCLFMT struct {
	gh      *github.Client
	ghToken string
	org     string
}

func (t HCLFMT) Name() string {
	return "hclfmt"
}

func (t HCLFMT) Desc() string {
	return "Apply hclfmt"
}

func (t HCLFMT) Apply(path string) (string, error) {
	var files []string
	if err := filepath.Walk(path, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error on %s: %s", file, err)
			return nil
		}
		if !skip(file, info) && isNomadFile(file) {
			log.Printf("%s: %s added to list", t.Name(), path)
			files = append(files, file)
		}
		return nil
	}); err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", nil
	}
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", errors.Wrap(err, "open repository")
	}
	log.Printf("git branch")
	branchName := branchNameFromDesc(t.Desc())
	if err := changeBranch(repo, branchName); err != nil {
		return "", errors.Wrapf(err, "change branch to %s", branchName)
	}

	for _, f := range files {
		log.Printf("Formatting %s", f)
		formatted, err := exec.Command("hclfmt", f).Output()
		if err != nil {
			return "", err
		}

		content, err := ioutil.ReadFile(f)
		if err != nil {
			return "", err
		}

		if string(content) == string(formatted) {
			continue
		}
		log.Printf("Formatted %s", f)
		if err := ioutil.WriteFile(f, formatted, 0644); err != nil {
			return "", err
		}
	}

	return handleGit(t.gh, t.ghToken, t.org, path, t.Desc())
}

func NewHCLFMT(gh *github.Client, ghToken, org string) HCLFMT {
	return HCLFMT{
		gh:      gh,
		org:     org,
		ghToken: ghToken,
	}
}
