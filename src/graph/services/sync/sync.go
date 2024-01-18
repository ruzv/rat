package sync

import (
	"context"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
	"rat/graph/services"
	"rat/logr"
)

var _ services.Service = (*Syncer)(nil)

// Config defines configuration params for periodically syncing graph to a
// git repository.
type Config struct {
	RepoDir            string        `yaml:"repoDir" validate:"nonzero"`
	Interval           time.Duration `yaml:"interval" validate:"nonzero"`
	PrivateKeyPath     string        `yaml:"privateKeyPath" validate:"nonzero"`
	PrivateKeyPassword string        `yaml:"privateKeyPassword"`
	KnownHostsPath     string        `yaml:"knownHostsPath" validate:"nonzero"`
}

// Syncer is a git syncer.
type Syncer struct {
	log           *logr.LogR
	interval      time.Duration
	repo          *git.Repository
	auth          *ssh.PublicKeys
	worktree      *git.Worktree
	trigger, stop chan struct{}
	lock          sync.Mutex
}

// NewSyncer creates a new Syncer.
func NewSyncer(
	c *Config, log *logr.LogR,
) (*Syncer, error) {
	log = log.Prefix("syncer")

	var (
		err error
		s   = &Syncer{
			log:      log,
			interval: c.Interval,
			trigger:  make(chan struct{}),
			stop:     make(chan struct{}),
		}
	)

	s.auth, err = ssh.NewPublicKeysFromFile(
		"git", c.PrivateKeyPath, c.PrivateKeyPassword,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth")
	}

	callback, err := ssh.NewKnownHostsCallback(c.KnownHostsPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create known hosts callback")
	}

	s.auth.HostKeyCallbackHelper.HostKeyCallback = callback

	s.repo, err = git.PlainOpen(c.RepoDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open repository")
	}

	s.worktree, err = s.repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get worktree")
	}

	return s, nil
}

// Start starts the sync ticker and goroutine.
func (s *Syncer) Run() error {
	ticker := time.NewTicker(s.interval)

	for {
		var err error
		select {
		case <-ticker.C:
			err = s.sync()
		case <-s.trigger:
			err = s.sync()
		case <-s.stop:
			ticker.Stop()
			close(s.trigger)
			close(s.stop)

			return nil
		}

		if err != nil {
			s.log.Errorf("failed to sync graph: %s", err.Error())
		}
	}
}

// Stop stops the sync ticker and goroutine, cleans up allocated resources.
func (s *Syncer) Stop(_ context.Context) error {
	// NOTE: stop should wait for sync to finish.
	s.log.Infof("stopping syncer")

	s.stop <- struct{}{}

	return nil
}

// Trigger triggers a sync.
func (s *Syncer) Trigger() {
	s.trigger <- struct{}{}
}

func (s *Syncer) sync() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	status, err := s.worktree.Status()
	if err != nil {
		return errors.Wrap(err, "failed to get worktree status")
	}

	if status.IsClean() {
		err = s.worktree.Pull(&git.PullOptions{Auth: s.auth})
		if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return errors.Wrap(err, "failed to pull changes")
		}

		return nil
	}

	err = s.worktree.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return errors.Wrap(err, "failed to git add all")
	}

	_, err = s.worktree.Commit(
		"rat sync",
		&git.CommitOptions{
			Author: &object.Signature{
				Name:  "rat sync client",
				Email: "zvejs.rudolfs@gmail.com",
				When:  time.Now(),
			},
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to commit changes")
	}

	err = s.worktree.Pull(&git.PullOptions{Auth: s.auth})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return errors.Wrap(err, "failed to pull changes")
	}

	err = s.repo.Push(&git.PushOptions{Auth: s.auth})
	if err != nil {
		return errors.Wrap(err, "failed to push changes to remote")
	}

	return nil
}
