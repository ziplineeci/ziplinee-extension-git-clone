package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/rs/zerolog/log"
	foundation "github.com/ziplineeci/ziplinee-foundation"
)

var (
	appgroup  string
	app       string
	version   string
	branch    string
	revision  string
	buildDate string
	goVersion = runtime.Version()
)

var (
	// flags
	gitSource            = kingpin.Flag("git-source", "The source of the repository.").Envar("ZIPLINEE_GIT_SOURCE").Required().String()
	gitOwner             = kingpin.Flag("git-owner", "The owner of the repository.").Envar("ZIPLINEE_GIT_OWNER").Required().String()
	gitName              = kingpin.Flag("git-name", "The owner plus repository name.").Envar("ZIPLINEE_GIT_NAME").Required().String()
	gitBranch            = kingpin.Flag("git-branch", "The branch to clone.").Envar("ZIPLINEE_GIT_BRANCH").Required().String()
	gitRevision          = kingpin.Flag("git-revision", "The revision to check out.").Envar("ZIPLINEE_GIT_REVISION").String()
	shallowClone         = kingpin.Flag("shallow-clone", "Shallow clone git repository for improved clone time.").Default("true").OverrideDefaultFromEnvar("ZIPLINEE_EXTENSION_SHALLOW").Bool()
	shallowCloneDepth    = kingpin.Flag("shallow-clone-depth", "Depth for shallow clone git repository for improved clone time.").Default("50").OverrideDefaultFromEnvar("ZIPLINEE_EXTENSION_DEPTH").Int()
	overrideRepo         = kingpin.Flag("override-repo", "Set other repository name to clone from same owner.").Envar("ZIPLINEE_EXTENSION_REPO").String()
	overrideBranch       = kingpin.Flag("override-branch", "Set other repository branch to clone from same owner.").Envar("ZIPLINEE_EXTENSION_BRANCH").String()
	overrideRevision     = kingpin.Flag("override-revision", "Set other repository revision to clone from same owner.").Envar("ZIPLINEE_EXTENSION_REVISION").String()
	overrideSubdirectory = kingpin.Flag("override-directory", "Set other repository directory to clone from same owner.").Envar("ZIPLINEE_EXTENSION_SUBDIR").String()

	bitbucketAPITokenPath = kingpin.Flag("bitbucket-api-token-path", "Path to file with Bitbucket api token credentials configured at the CI server, passed in to this trusted extension.").Default("/credentials/bitbucket_api_token.json").String()
	githubAPITokenPath    = kingpin.Flag("github-api-token-path", "Path to file with Github api token credentials configured at the CI server, passed in to this trusted extension.").Default("/credentials/github_api_token.json").String()
)

func main() {

	// parse command line parameters
	kingpin.Parse()

	// init log format from envvar ZIPLINEE_LOG_FORMAT
	applicationInfo := foundation.ApplicationInfo{
		AppGroup:  appgroup,
		App:       app,
		Version:   version,
		Branch:    branch,
		Revision:  revision,
		BuildDate: buildDate,
	}
	foundation.InitLoggingFromEnv(applicationInfo)

	// create context to cancel commands on sigterm
	ctx := foundation.InitCancellationContext(context.Background())

	// get api token from injected credentials
	bitbucketAPIToken := ""
	// use mounted credential file if present instead of relying on an envvar
	if runtime.GOOS == "windows" {
		*bitbucketAPITokenPath = "C:" + *bitbucketAPITokenPath
	}
	if foundation.FileExists(*bitbucketAPITokenPath) {
		var credentials []APITokenCredentials
		log.Info().Msgf("Reading credentials from file at path %v...", *bitbucketAPITokenPath)
		credentialsFileContent, err := ioutil.ReadFile(*bitbucketAPITokenPath)
		if err != nil {
			log.Fatal().Msgf("Failed reading credential file at path %v.", *bitbucketAPITokenPath)
		}
		err = json.Unmarshal(credentialsFileContent, &credentials)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed unmarshalling injected credentials")
		}
		bitbucketAPIToken = credentials[0].AdditionalProperties.Token
	}

	githubAPIToken := ""
	// use mounted credential file if present instead of relying on an envvar
	if runtime.GOOS == "windows" {
		*githubAPITokenPath = "C:" + *githubAPITokenPath
	}
	if foundation.FileExists(*githubAPITokenPath) {
		var credentials []APITokenCredentials
		log.Info().Msgf("Reading credentials from file at path %v...", *githubAPITokenPath)
		credentialsFileContent, err := ioutil.ReadFile(*githubAPITokenPath)
		if err != nil {
			log.Fatal().Msgf("Failed reading credential file at path %v.", *githubAPITokenPath)
		}
		err = json.Unmarshal(credentialsFileContent, &credentials)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed unmarshalling injected credentials")
		}
		githubAPIToken = credentials[0].AdditionalProperties.Token
	}

	if *overrideRepo != "" {
		if *overrideBranch == "" {
			*overrideBranch = "master"
		}
		if *overrideSubdirectory == "" {
			*overrideSubdirectory = *overrideRepo
		}

		overrideGitURL := fmt.Sprintf("https://%v/%v/%v", *gitSource, *gitOwner, *overrideRepo)
		if strings.HasPrefix(*overrideRepo, "https://") {
			// this allows for any public repo to be cloned
			overrideGitURL = *overrideRepo
		} else {
			if bitbucketAPIToken != "" {
				overrideGitURL = fmt.Sprintf("https://x-token-auth:%v@%v/%v/%v", bitbucketAPIToken, *gitSource, *gitOwner, *overrideRepo)
			}
			if githubAPIToken != "" {
				overrideGitURL = fmt.Sprintf("https://x-access-token:%v@%v/%v/%v", githubAPIToken, *gitSource, *gitOwner, *overrideRepo)
			}
		}

		// git clone the specified repository branch to the specific directory
		err := gitCloneOverride(ctx, *overrideRepo, overrideGitURL, *overrideBranch, *overrideRevision, *overrideSubdirectory, *shallowClone, *shallowCloneDepth)
		if err != nil {
			log.Fatal().Err(err).Msgf("Error cloning git repository %v to branch %v into subdir %v", *overrideRepo, *overrideBranch, *overrideSubdirectory)
		}

		return
	}

	gitURL := fmt.Sprintf("https://%v/%v/%v", *gitSource, *gitOwner, *gitName)
	if bitbucketAPIToken != "" {
		gitURL = fmt.Sprintf("https://x-token-auth:%v@%v/%v/%v", bitbucketAPIToken, *gitSource, *gitOwner, *gitName)
		ctx = context.WithValue(ctx, "source", "bitbucket")
		ctx = context.WithValue(ctx, "token", bitbucketAPIToken)
	}
	if githubAPIToken != "" {
		gitURL = fmt.Sprintf("https://x-access-token:%v@%v/%v/%v", githubAPIToken, *gitSource, *gitOwner, *gitName)
		ctx = context.WithValue(ctx, "source", "github")
		ctx = context.WithValue(ctx, "token", githubAPIToken)
	}

	// git clone to specific branch and revision
	err := gitCloneRevision(ctx, *gitName, gitURL, *gitBranch, *gitRevision, *shallowClone, *shallowCloneDepth)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error cloning git repository %v to branch %v and revision %v with shallow clone is %v", *gitName, *gitBranch, *gitRevision, *shallowClone)
	}
}
