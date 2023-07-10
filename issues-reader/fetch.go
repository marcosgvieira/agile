package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"os"
	"strconv"
	"strings"
)

const RepoOwner = "marcosgvieira"
const CloudRepoOwner = "camunda-cloud"
const MainRepoName = "agile"
const ZeebeRepoName = "zeebe"
const OperateRepoName = "operate"
const TasklistRepoName = "tasklist"
const IdentityRepoName = "identity"
const ReleaseNotesTemplateFileName = "release-notes-template.txt"

type CamundaPlatformRelease struct {
	ZeebeReleaseNotes    string
	OperateReleaseNotes  string
	TasklistReleaseNotes string
	IdentityReleaseNotes string
}


func listCommitsBetweenTags(ctx context.Context, client *github.Client, owner, repo, tag1, tag2 string) ([]*github.RepositoryCommit, error) {

	log.Debug().Msg("owner = " + owner + " repo = " + repo + " tag1 = " + tag1 + "tag2 = " + tag2)

	// Retrieve the commit range between the tags
	commits, _, err := client.Repositories.CompareCommits(ctx, owner, repo, "111", "222", &github.ListOptions{
	})
	if err != nil {
		return nil, err
	}
	for _, commit := range commits.Commits {
		message := *commit.Commit.Message
		log.Debug().Msg("Message = " + message)
	}
	return commits.Commits, nil
}



// findClosedIssues retrieves the closed GitHub issues associated with the given commits.
func findClosedIssues(ctx context.Context, client *github.Client, owner, repo string, commits []*github.RepositoryCommit) ([]*github.Issue, error) {
	// Initialize a set to store the closed issue numbers
	closedIssues := make(map[int]bool)

	// Iterate over the commits and retrieve the associated pull requests
	for _, commit := range commits {
		// Get the SHA of the commit
		sha := *commit.SHA

		// Retrieve the pull requests associated with the commit
		pulls, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
			State: "closed",
			Head:  sha,
		})
		if err != nil {
			return nil, err
		}

		// Iterate over the pull requests and retrieve the closed issue numbers
		for _, pull := range pulls {
			issueURL := pull.GetIssueURL()
			issueNumber, err := extractIssueNumberFromURL(issueURL)
			if err != nil {
				return nil, err
			}

			log.Debug().Msg("links= " + strconv.Atoi(issueNumber))
			// Retrieve the issue details
			issue, _, err := client.Issues.Get(ctx, owner, repo, issueNumber)
			if err != nil {
				return nil, err
			}

			if(issue != nil) {
				log.Debug().Msg("so por aqui")
			}

		}
	}

	// Retrieve the closed issue details
	issues := []*github.Issue{}
	for issueNumber := range closedIssues {
		// Retrieve the closed issue
		issue, _, err := client.Issues.Get(ctx, owner, repo, issueNumber)
		if err != nil {
			return nil, err
		}

		issues = append(issues, issue)
	}

	return issues, nil
}

// Extract the issue number from the issue URL
func extractIssueNumberFromURL(url string) (int, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid issue URL")
	}
	issueNumber, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0, err
	}
	return issueNumber, nil
}





func main() {

	camundaTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_CAMUNDA_ACCESS_TOKEN")},
	)

	ctx := context.TODO()
	camundaOAuthClient := oauth2.NewClient(ctx, camundaTokenSource)

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	camundaGithubClient := github.NewClient(camundaOAuthClient)

    var githubRef = "222"

	log.Debug().Msg("Github ref = " + githubRef)

	commits, err := listCommitsBetweenTags(
		ctx,
		camundaGithubClient,
		RepoOwner,
		"agile",
		"111",
		"222",
		)
	if err != nil {
		log.Debug().Msg("error = " + err.Error())
	}

	issues, err := findClosedIssues(ctx, camundaGithubClient, RepoOwner, "agile", commits)
	if err != nil {
		log.Debug().Msg("error = " + err.Error())
	}

	if(issues != nil){
		log.Debug().Msg("issues = ")
	}

	if(commits != nil) {
		log.Debug().Msg("commits = ")
	}

}
