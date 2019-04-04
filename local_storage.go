package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"gopkg.in/src-d/go-git.v4"
)

type repository struct {
	path string
	*github.Repository
}

func (r *repository) Path() string {
	return r.path
}

func isValidRepo(path string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return errors.Wrap(err, "git open")
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return errors.Wrapf(err, "get %s remotes", path)
	}
	if len(remotes) == 0 {
		return fmt.Errorf("no remotes")
	}
	for _, remote := range remotes {
		spew.Dump(remote)
		for _, url := range remote.Config().URLs {
			_ = url
		}
	}

	return nil
}

func gatherLocalRepos(localStoragePath string) ([]string, error) {
	repos := []string{}
	dirs, err := ioutil.ReadDir(localStoragePath)
	if err != nil {
		return nil, errors.Wrapf(err, "list %s", localStoragePath)
	}
	for _, f := range dirs {
		if !f.IsDir() {
			continue
		}
		p := path.Join(localStoragePath, f.Name())
		if err := isValidRepo(p); err != nil {
			log.Println(errors.Wrapf(err, "check %s validity", p))
			continue
		}
		repos = append(repos, p)
	}

	return repos, nil
}

func listDirs(path string) ([]string, error) {
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, errors.Wrapf(err, "list %s", path)
	}

	dirs := []string{}
	for _, f := range infos {
		if !f.IsDir() {
			continue
		}
		dirs = append(dirs, filepath.Join(path, f.Name()))
	}
	return dirs, nil
}

func inList(list []string, wanted string) bool {
	for _, el := range list {
		if filepath.Base(el) == wanted {
			return true
		}
	}

	return false
}

func fetchRepo(localStoragePath string, repo *github.Repository) error {
	return nil
}

func getRepos(repos []*github.Repository, localStoragePath string) ([]repository, error) {
	var od *github.Repository
	for _, r := range repos {
		if *r.Name != "one-delivery" {
			continue
		}
		od = r
		break
	}
	return []repository{
		repository{
			path:       filepath.Join(localStoragePath, *od.Name),
			Repository: od,
		},
	}, nil

	// dirs, err := listDirs(localStoragePath)
	// if err != nil {
	// 	return nil, err
	// }

	// toFetch := []*github.Repository{}
	// for _, repo := range repos {
	// 	if !inList(dirs, *repo.Name) {
	// 		toFetch = append(toFetch, repo)
	// 		continue
	// 	}
	// }

	// log.Printf("Fetching %d repositories", len(toFetch))
	// for _, repo := range toFetch {
	// 	if err := fetchRepo(localStoragePath, repo); err != nil {
	// 		return nil, err
	// 	}
	// }

	// fetched := []repository{}
	// for _, repo := range repos {
	// 	fetched = append(fetched, repository{
	// 		path:       filepath.Join(localStoragePath, *repo.Name),
	// 		Repository: repo,
	// 	})
	// }

	// return fetched, nil
}
