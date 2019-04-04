package main

import (
	"context"
	"log"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func gatherOrgRepos(ctx context.Context, client *github.Client, org string) ([]*github.Repository, error) {
	log.Printf("Gathering repositories from %s", org)

	repos := []*github.Repository{}
	opts := github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 1000},
	}
	for {
		rs, resp, err := client.Repositories.ListByOrg(ctx, org, &opts)
		if err != nil {
			return nil, err
		}
		repos = append(repos, rs...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	log.Printf("Got %d repositories\n", len(repos))
	return repos, nil
}

func newClient(ctx context.Context, accessToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
}
