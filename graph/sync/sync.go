package sync

import (
	"sync"
	"time"

	"rat/config"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
)

var log = logging.MustGetLogger("sync")

type Syncer struct {
	interval      time.Duration
	repo          *git.Repository
	auth          *ssh.PublicKeys
	worktree      *git.Worktree
	trigger, stop chan struct{}
	lock          sync.Mutex
}

// NewSyncer creates a new Syncer.
func NewSyncer(repoDir string, conf *config.SyncConfig) (*Syncer, error) {
	var (
		err error
		s   = &Syncer{
			interval: conf.Interval,
			trigger:  make(chan struct{}),
			stop:     make(chan struct{}),
		}
	)

	s.repo, err = git.PlainOpen(repoDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open repository")
	}

	s.auth, err = ssh.NewPublicKeysFromFile(
		"git", conf.KeyPath, conf.KeyPassword,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create auth")
	}

	s.worktree, err = s.repo.Worktree()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get worktree")
	}

	return s, nil
}

func (s *Syncer) Start() {
	ticker := time.NewTicker(s.interval)

	go func() {
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

				return
			}

			if err != nil {
				log.Errorf("failed to sync graph: %s", err.Error())
			}
		}
	}()
}

func (s *Syncer) Stop() {
	s.stop <- struct{}{}
}

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
