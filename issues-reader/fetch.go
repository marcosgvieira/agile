package main

import (
	"context"
	"github.com/google/go-github/v51/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"os"
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


func findLinkedIssues(ctx context.Context, client *github.Client, owner, repo string, commits []*github.RepositoryCommit) ([]*github.Issue, error) {
	// Initialize the list of linked issues
	issues := []*github.Issue{}

	for _, commit := range commits {
		// Get the SHA of the commit
		sha := *commit.SHA

		// Retrieve the pull requests associated with the commit
		pulls, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, sha, nil)
		if err != nil {
			return nil, err
		}

		// Iterate over the pull requests and extract the linked issues
		for _, pull := range pulls {
			linkedIssues, err := findClosedIssues(ctx, client, owner, repo, *pull.Number)
			if err != nil {
				return nil, err
			}

			issues = append(issues, linkedIssues...)
		}
	}

	return issues, nil
}



func findClosedIssues(ctx context.Context, client *github.Client, owner, repo string, pullNumber int) ([]*github.Issue, error) {
	// Retrieve the events for the pull request
	events, _, err := client.Issues.ListIssueEvents(ctx, owner, repo, pullNumber, nil)
	if err != nil {
		return nil, err
	}

	// Initialize a set to store the closed issue numbers
	closedIssues := make(map[int]bool)

	// Iterate over the events and check for closed issue events
	for _, event := range events {
		if event.Event == "closed" && event.Issue != nil {
			closedIssues[*event.Issue.Number] = true
		}
	}

	// Retrieve the closed issue details
	issues := []*github.Issue{}
	for closedIssue := range closedIssues {
		issue, _, err := client.Issues.Get(ctx, owner, repo, closedIssue)
		if err != nil {
			return nil, err
		}

		issues = append(issues, issue)
	}

	return issues, nil
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

	issues, err := findLinkedIssues(ctx, camundaGithubClient, RepoOwner, "agile", commits)
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
