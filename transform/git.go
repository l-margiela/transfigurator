package transform

import (
	"context"
	"log"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"github.com/google/go-github/github"

	"gopkg.in/src-d/go-git.v4/plumbing"

	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"gopkg.in/src-d/go-git.v4"

	"github.com/pkg/errors"
)

func gitChanges(repo *git.Repository) (bool, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return false, err
	}

	status, err := wt.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean(), nil
}

func changeBranch(repo *git.Repository, branchName string) error {

	head, err := repo.Head()
	if err != nil {
		return err
	}

	ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/"+branchName), head.Hash())
	if err := repo.Storer.SetReference(ref); err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	if err := wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/" + branchName),
	}); err != nil {
		return errors.Wrapf(err, "checkout %s", branchName)
	}

	return nil
}

func commit(repo *git.Repository, msg string) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	if _, err := wt.Commit(msg, &git.CommitOptions{
		All: true,
		Author: &object.Signature{
			Name:  "Platform Engineering",
			Email: "platform-engineering@bydeluxe.com",
			When:  time.Now(),
		},
	}); err != nil {
		return err
	}

	return nil
}

func makePR(client *github.Client, org string, repoName, branch, title, body string) (string, error) {
	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(branch),
		Base:                github.String("master"),
		Body:                github.String(body),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(context.Background(), org, repoName, newPR)
	if err != nil {
		return "", err
	}
	return *pr.HTMLURL, nil
}

func branchNameFromDesc(desc string) string {
	title := strings.SplitN(desc, "\n", 1)[0]
	branchName := strings.Map(func(in rune) rune {
		if in == ' ' {
			return '-'
		}
		return in
	}, title)
	log.Printf(branchName)
	return strings.ToLower(branchName)
}

func handleGit(gh *github.Client, ghToken, org, path, gitMsg string) (string, error) {
	log.Printf("git open")
	repo, err := git.PlainOpen(path)
	if err != nil {
		return "", errors.Wrap(err, "open repository")
	}

	log.Printf("git changes")
	if hasChanges, err := gitChanges(repo); err != nil {
		return "", err
	} else if !hasChanges {
		log.Printf("no changes")
		return "", nil
	}
	wt, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	status, err := wt.Status()
	if err != nil {
		return "", err
	}
	log.Println(status)

	log.Printf("git commit")
	if err := commit(repo, gitMsg); err != nil {
		return "", errors.Wrap(err, "commit git changes")
	}

	log.Printf("git push")
	if err := repo.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "xaxes",
			Password: ghToken,
		},
	}); err != nil {
		return "", errors.Wrap(err, "git push")
	}

	log.Printf("git PR")
	desc := strings.Split(gitMsg, "\n")
	prTitle := desc[0]
	prBody := strings.Join(desc[1:], "\n")
	branchName := branchNameFromDesc(gitMsg)
	return makePR(gh, org, filepath.Base(path), branchName, prTitle, prBody)
}
