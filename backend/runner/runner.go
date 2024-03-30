package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
	"woodpecker-ci/logger"

	"github.com/drone/envsubst"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/docker"
	backendTypes "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/linter"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/matrix"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

type RunnerConfigOptions struct {
	NoProxy       bool              `json:"no_proxy"`
	HTTPProxy     string            `json:"http_proxy"`
	HTTPSProxy    string            `json:"https_proxy"`
	WorkspaceBase string            `json:"workspace_base"` // /home/woody/repos
	WorkspacePath string            `json:"workspace_path"` // project-name
	Env           []string          `json:"env"`
	Privileged    []string          `json:"privileged"`
	Secrets       map[string]string `json:"secrets"`
	Timeout       time.Duration     `json:"timeout"`
}

// the cli.Context need to be replaced by flags
func RunExec(c *cli.Context, conf RunnerConfigOptions, file, repoPath string) error {
	log := logger.GetLogger()
	dat, err := os.ReadFile(file)
	if err != nil {
		log.WithField("file", file).Error(err)
		return err
	}

	axes, err := matrix.ParseString(string(dat))
	if err != nil {
		log.WithField("file", file).Error(err)
		return fmt.Errorf("parse matrix fail")
	}

	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}
	for _, axis := range axes {
		err := execWithAxis(c, conf, file, repoPath, axis)
		if err != nil {
			log.WithFields(logrus.Fields{
				"file":     file,
				"config":   conf,
				"repoPath": repoPath,
				"axis":     axis,
			}).WithError(err).Error("failed to exec axis")
			return err
		}
	}
	return nil
}

func execWithAxis(c *cli.Context, config RunnerConfigOptions, file, repoPath string, axis matrix.Axis) error {
	logg := logger.GetLogger()
	metadata := metadataFromContext(c, axis)
	environ := metadata.Environ()
	var secrets []compiler.Secret
	// here we only need the axis metadata, above functions calls could almost be removed
	for key, val := range metadata.Workflow.Matrix {
		environ[key] = val
		secrets = append(secrets, compiler.Secret{
			Name:  key,
			Value: val,
		})
	}

	for key, val := range config.Secrets {
		secrets = append(secrets, compiler.Secret{
			Name:  key,
			Value: val,
		})
	}

	pipelineEnv := make(map[string]string)
	for _, env := range config.Env {
		before, after, _ := strings.Cut(env, "=")
		pipelineEnv[before] = after
		if oldVar, exists := environ[before]; exists {
			// override existing values, but print a warning
			log.Warn().Msgf("environment variable '%s' had value '%s', but got overwritten", before, oldVar)
		}
		environ[before] = after
	}

	tmpl, err := envsubst.ParseFile(file)
	if err != nil {
		logg.WithError(err).Error("failed to parse file")
		return err
	}
	confstr, err := tmpl.Execute(func(name string) string {
		return environ[name]
	})
	if err != nil {
		logg.WithError(err).Error("failed to execute file")
		return err
	}

	conf, err := yaml.ParseString(confstr)
	if err != nil {
		logg.WithError(err).Error("failed to parse yaml string")
		return err
	}

	// configure volumes for local execution
	//volumes := c.StringSlice("volumes")

	//specific to local mode
	var (
		workspaceBase = conf.Workspace.Base
		workspacePath = conf.Workspace.Path
	)
	if workspaceBase == "" {
		workspaceBase = config.WorkspaceBase
	}
	if workspacePath == "" {
		workspacePath = config.WorkspacePath
	}

	volumes := []string{}
	prefix := "abcd"
	volumes = append(volumes, prefix+"_default:"+workspaceBase)
	volumes = append(volumes, repoPath+":"+path.Join(workspaceBase, workspacePath))
	//end of specific to local mode

	// lint the yaml file
	if lerr := linter.New(linter.WithTrusted(true)).Lint([]*linter.WorkflowConfig{{
		File:      path.Base(file),
		RawConfig: confstr,
		Workflow:  conf,
	}}); lerr != nil {
		logg.WithError(lerr).Error("failed to lint yaml string")
		return lerr
	}

	// compiles the yaml file
	compiled, err := compiler.New(
		compiler.WithEscalated(
			config.Privileged...,
		),
		compiler.WithVolumes(volumes...),
		compiler.WithWorkspace(
			config.WorkspaceBase,
			config.WorkspacePath,
		),
		/*compiler.WithNetworks(
			c.StringSlice("network")...,
		),*/
		/*compiler.WithPrefix(
			c.String("prefix"),
		),*/

		/*compiler.WithProxy(compiler.ProxyOptions{
			NoProxy:    c.String("backend-no-proxy"),
			HTTPProxy:  c.String("backend-http-proxy"),
			HTTPSProxy: c.String("backend-https-proxy"),
		}),*/

		compiler.WithLocal(true),
		/*compiler.WithNetrc(
			c.String("netrc-username"),
			c.String("netrc-password"),
			c.String("netrc-machine"),
		),*/
		compiler.WithMetadata(metadata),
		compiler.WithSecret(secrets...),
		compiler.WithEnviron(pipelineEnv),
	).Compile(conf)
	if err != nil {
		logg.WithError(err).Error("failed to compile pipeline")
		return err
	}

	backendCtx := context.WithValue(c.Context, backendTypes.CliContext, c)
	backendEngine := docker.New() // 		local.New(),

	if _, err = backendEngine.Load(backendCtx); err != nil {
		logg.WithError(err).Error("failed to load backend engine")
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()
	ctx = utils.WithContextSigtermCallback(ctx, func() {
		fmt.Println("ctrl+c received, terminating process")
	})

	p := pipeline.New(compiled,
		pipeline.WithContext(ctx),
		pipeline.WithTracer(pipeline.DefaultTracer), // need to provide a custom tracer here. the default one use env var
		pipeline.WithLogger(defaultLogger),
		pipeline.WithBackend(backendEngine),
		pipeline.WithDescription(map[string]string{
			"CLI": "exec",
		}),
	)

	return p.Run(c.Context)
}

var defaultLogger = pipeline.Logger(func(step *backendTypes.Step, rc io.Reader) error {
	_, err := io.Copy(os.Stdout, rc)
	return err
})

func metadataFromContext(c *cli.Context, axis matrix.Axis) metadata.Metadata {
	platform := c.String("system-platform")
	if platform == "" {
		platform = runtime.GOOS + "/" + runtime.GOARCH
	}

	fullRepoName := c.String("repo-name")
	repoOwner := ""
	repoName := ""
	if idx := strings.LastIndex(fullRepoName, "/"); idx != -1 {
		repoOwner = fullRepoName[:idx]
		repoName = fullRepoName[idx+1:]
	}

	return metadata.Metadata{
		Repo: metadata.Repo{
			Name:        repoName,
			Owner:       repoOwner,
			RemoteID:    c.String("repo-remote-id"),
			ForgeURL:    c.String("repo-url"),
			CloneURL:    c.String("repo-clone-url"),
			CloneSSHURL: c.String("repo-clone-ssh-url"),
			Private:     c.Bool("repo-private"),
			Trusted:     c.Bool("repo-trusted"),
		},
		Curr: metadata.Pipeline{
			Number:   c.Int64("pipeline-number"),
			Parent:   c.Int64("pipeline-parent"),
			Created:  c.Int64("pipeline-created"),
			Started:  c.Int64("pipeline-started"),
			Finished: c.Int64("pipeline-finished"),
			Status:   c.String("pipeline-status"),
			Event:    c.String("pipeline-event"),
			ForgeURL: c.String("pipeline-url"),
			Target:   c.String("pipeline-target"),
			Commit: metadata.Commit{
				Sha:     c.String("commit-sha"),
				Ref:     c.String("commit-ref"),
				Refspec: c.String("commit-refspec"),
				Branch:  c.String("commit-branch"),
				Message: c.String("commit-message"),
				Author: metadata.Author{
					Name:   c.String("commit-author-name"),
					Email:  c.String("commit-author-email"),
					Avatar: c.String("commit-author-avatar"),
				},
			},
		},
		Prev: metadata.Pipeline{
			Number:   c.Int64("prev-pipeline-number"),
			Created:  c.Int64("prev-pipeline-created"),
			Started:  c.Int64("prev-pipeline-started"),
			Finished: c.Int64("prev-pipeline-finished"),
			Status:   c.String("prev-pipeline-status"),
			Event:    c.String("prev-pipeline-event"),
			ForgeURL: c.String("prev-pipeline-url"),
			Commit: metadata.Commit{
				Sha:     c.String("prev-commit-sha"),
				Ref:     c.String("prev-commit-ref"),
				Refspec: c.String("prev-commit-refspec"),
				Branch:  c.String("prev-commit-branch"),
				Message: c.String("prev-commit-message"),
				Author: metadata.Author{
					Name:   c.String("prev-commit-author-name"),
					Email:  c.String("prev-commit-author-email"),
					Avatar: c.String("prev-commit-author-avatar"),
				},
			},
		},
		Workflow: metadata.Workflow{
			Name:   c.String("workflow-name"),
			Number: c.Int("workflow-number"),
			Matrix: axis,
		},
		Step: metadata.Step{
			Name:   c.String("step-name"),
			Number: c.Int("step-number"),
		},
		Sys: metadata.System{
			Name:     c.String("system-name"),
			URL:      c.String("system-url"),
			Platform: platform,
			Version:  version.Version,
		},
		Forge: metadata.Forge{
			Type: c.String("forge-type"),
			URL:  c.String("forge-url"),
		},
	}
}
