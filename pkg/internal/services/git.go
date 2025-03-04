package services

import (
	"context"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	"github.com/pkg/errors"
	"io"
	"net/url"
	"strings"
)

type (
	CloneOptions struct {
		allowCommit bool
	}
	GitService interface {
		Clone(
			ctx context.Context,
			gitUrl string,
			gitBranch string,
			optionsFn ...func(*CloneOptions),
		) (*git.Repository, error)
		ReadFile(
			repo *git.Repository,
			path string,
		) ([]byte, error)
		CommitFile(
			repo *git.Repository,
			path string,
			body []byte,
			commitMessage string,
		) error
		OpenDir(
			repo *git.Repository,
			path string,
		) (*object.Tree, error)
		CopyRepo(
			ctx context.Context,
			srcGitUrl, dstGitUrl string,
		) error
	}
)

const (
	ServiceKeyGitService = "gitService"
)

func NewGitService() GitService {
	return &gitService{}
}

func extractTokenFromURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	if parsedURL.User != nil {
		password, exists := parsedURL.User.Password()
		if exists {
			return password
		}
		return parsedURL.User.Username()
	}
	return ""
}

type gitService struct{}

func (s *gitService) Clone(
	ctx context.Context,
	gitUrl string,
	gitBranch string,
	optionsFn ...func(*CloneOptions),
) (*git.Repository, error) {
	var options CloneOptions
	for _, fn := range optionsFn {
		fn(&options)
	}

	// Extract token from the gitUrl
	token := extractTokenFromURL(gitUrl)

	var auth transport.AuthMethod
	if token != "" {
		auth = &http.BasicAuth{
			Username: "git",
			Password: token,
		}
	}

	cloneOptions := git.CloneOptions{
		URL:           gitUrl,
		ReferenceName: plumbing.NewBranchReferenceName(gitBranch),
		SingleBranch:  true,
		Auth:          auth,
	}

	var w billy.Filesystem
	if options.allowCommit {
		w = memfs.New()
	}

	logger.Debug("Cloning git repo..", "url", gitUrl, "branch", gitBranch)
	repo, err := git.CloneContext(ctx, memory.NewStorage(), w, &cloneOptions)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to clone git repo")
	}
	return repo, nil
}

func (s *gitService) ReadFile(
	repo *git.Repository,
	path string,
) ([]byte, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get head")
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get commit object")
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get tree")
	}

	file, err := tree.File(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get file: %s", path)
	}

	reader, err := file.Reader()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get reader")
	}

	contents, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read all")
	}

	return contents, nil
}

func (s *gitService) OpenDir(
	repo *git.Repository,
	path string,
) (*object.Tree, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get head")
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get commit object")
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get tree")
	}

	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return tree, nil
	}

	path = strings.TrimPrefix(path, "/")
	dirTree, err := tree.Tree(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dir. path='%s'", path)
	}

	return dirTree, nil
}

func (s *gitService) CommitFile(
	repo *git.Repository,
	path string,
	body []byte,
	commitMessage string,
) error {
	wt, err := repo.Worktree()
	if err != nil {
		return errors.Wrapf(err, "failed to get worktree")
	}

	fs := wt.Filesystem
	if err := util.WriteFile(fs, path, body, 0644); err != nil {
		return errors.Wrapf(err, "failed to write file")
	}

	w, err := repo.Worktree()
	if err != nil {
		return errors.Wrapf(err, "failed to get worktree")
	}

	if _, err = w.Add(path); err != nil {
		return errors.Wrapf(err, "failed to add")
	}

	_, err = w.Commit(commitMessage, &git.CommitOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to commit")
	}

	return nil
}

func (s *gitService) CopyRepo(
	ctx context.Context,
	sourceGitUrl string,
	remoteGitUrl string,
) error {
	repo, err := s.Clone(ctx, sourceGitUrl, "main", GitCloneOptionsAllowCommit())
	if err != nil {
		return errors.Wrapf(err, "failed to clone")
	}

	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "target",
		URLs: []string{remoteGitUrl},
	})
	if err != nil {
		return errors.Wrapf(err, "failed to create remote")
	}

	if err := repo.PushContext(ctx, &git.PushOptions{
		RemoteName: remote.Config().Name,
		RefSpecs: []config.RefSpec{
			"+refs/heads/main:refs/heads/main",
		},
		Force: true,
	}); err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return errors.Wrapf(err, "failed to push")
		}
		logger.Info("remote already up-to-date, no push done", "remote", remote.Config().Name)
	}

	logger.Debug("Force push to target completed successfully", "target", remote.Config().Name)
	return nil
}

func GitCloneOptionsAllowCommit() func(*CloneOptions) {
	return func(o *CloneOptions) {
		o.allowCommit = true
	}
}

func init() {
	digo.ProvideService(ServiceKeyGitService, func(ctx *digo.Container) (interface{}, error) {
		return NewGitService(), nil
	})
}
