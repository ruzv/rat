package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/pkg/errors"
)

func main() {
	repoDir := "./"
	pemFilePath := "/home/ruzv/.ssh/github-mech"

	err := gitSyncGraphChanges(repoDir, pemFilePath)
	if err != nil {
		panic(err)
	}
}

func gitSyncGraphChanges(repoDir, pemFilePath string) error {
	r, err := git.PlainOpen(repoDir)
	if err != nil {
		return errors.Wrap(err, "failed to open repository")
	}

	w, err := r.Worktree()
	if err != nil {
		return errors.Wrap(err, "failed to get worktree")
	}

	s, err := w.Status()
	if err != nil {
		return errors.Wrap(err, "failed to get worktree status")
	}

	if s.IsClean() {
		return nil
	}

	err = w.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return errors.Wrap(err, "failed to git add all")
	}

	_, err = w.Commit(
		"rat sync",
		&git.CommitOptions{
			Author: &object.Signature{
				Name:  "rat sync client",
				Email: "zvejs.rudolfs@gmail.com",
			},
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to commit changes")
	}

	auth, err := ssh.NewPublicKeysFromFile("git", pemFilePath, "")
	if err != nil {
		return errors.Wrap(err, "failed to create auth")
	}

	// change

	err = r.Push(&git.PushOptions{Auth: auth})
	if err != nil {
		return errors.Wrap(err, "failed to push changes to remote")
	}

	return nil
}
