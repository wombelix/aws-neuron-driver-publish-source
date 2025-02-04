// SPDX-FileCopyrightText: 2025 Dominik Wombacher <dominik@wombacher.cc>
//
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func getGitRepo(directory string) *git.Repository {
	repo, err := git.PlainOpen(directory)
	checkError(err)

	return repo
}

func getGitRepoWorktreeFromRepo(repo *git.Repository) *git.Worktree {
	worktree, err := repo.Worktree()
	checkError(err)

	return worktree
}

func getGitRepoWorktree(directory string) *git.Worktree {
	repo := getGitRepo(directory)

	worktree, err := repo.Worktree()
	checkError(err)

	return worktree
}

func gitWorktreeModified(directory string) bool {
	worktree := getGitRepoWorktree(directory)

	status, err := worktree.Status()
	checkError(err)

	return !status.IsClean()
}

func featureBranchCommitMerge(directory string, featureBranch string, commitMsg string) {
	var err error
	repo := getGitRepo(directory)
	worktree := getGitRepoWorktreeFromRepo(repo)

	referenceName := plumbing.NewBranchReferenceName(featureBranch)

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: referenceName,
		Create: true,
		Keep:   true,
	})
	checkError(err)

	worktreeStatus, err := worktree.Status()
	checkError(err)

	for path, status := range worktreeStatus {
		if status.Worktree != git.Unmodified {
			_, err = worktree.Add(path)
			checkError(err)

			commitMsg += fmt.Sprintf("- %s\n", path)
		}
	}

	_, err = worktree.Commit(commitMsg, &git.CommitOptions{
		All: true,
	})
	checkError(err)

	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.Main,
	})
	checkError(err)

	reference, err := repo.Reference(referenceName, false)
	checkError(err)

	err = repo.Merge(
		*reference,
		git.MergeOptions{},
	)
	checkError(err)

	// Additional step required to complete feature branch merge, same as in worktree.PullContext:
	// https://github.com/go-git/go-git/blob/main/worktree.go#L137
	err = worktree.Reset(&git.ResetOptions{
		Mode:   git.MergeReset,
		Commit: reference.Hash(),
	})
	checkError(err)

	// 'repo.DeleteBranch' doesn't do what the name says.
	// It needs 'repo.Storer.RemoveReference' to delete a branch.
	// https://stackoverflow.com/questions/55745430/deleting-local-branch-in-go-git-branch-not-found#comment118723622_55745430
	err = repo.Storer.RemoveReference(referenceName)
	checkError(err)
}
