package main

import (
	"fmt"
	"log"

	"github.com/go-git/go-git/v5"
	gitconf "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/ikaruswill/argonaut/internal/config"
)

func main() {
	// Prerequisites: Helm, Kustomize
	// Configs:
	// - Helm version
	// - Kustomize version (Optional)
	// - Argo CD token
	// - Gitlab token
	// - Instance profile for argocd-vault-plugin
	// - Gitlab CI current repo URL
	// - Master branch name (Default master)
	// - Gitlab CI current branch
	// - Gitlab CI current commit
	//
	// Logic flow
	// Load Config
	// Get git changed files
	// Sort change list by directory tree
	// For each changed directory
	// - Check if directory contains Application definitions
	// - Determine if Application is helm, kustomize or plain yaml
	// - For Kustomize apps
	//     - Run kustomize build
	// - For Helm apps
	//     - Run helm template from Application values key
	// - Check for any avp annotations
	//     - Apply argocd-vault-plugin generate
	// - Run argocd app diff
	// - Collate outputs
	// For each changed directory
	// - Format output
	// - Merge into single text body
	// Submit comment to GitLab API
	conf := config.Load()

	repoAuth := http.BasicAuth{
		Username: "git",
		Password: conf.GitLabToken,
	}

	// Cleanup working directory
	// os.RemoveAll("./")

	// repo, err := git.PlainClone("./", false, &git.CloneOptions{
	// 	Auth:     &repoAuth,
	// 	URL:      conf.RepoURL,
	// 	Progress: os.Stdout,
	// })
	repo, err := git.PlainOpen("./")
	if err != nil {
		log.Printf("Fail git.PlainClone %s", err.Error())
	}

	// Fetch all remotes
	repo.Fetch(&git.FetchOptions{
		Auth: &repoAuth,
		RefSpecs: []gitconf.RefSpec{
			gitconf.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", conf.Branch, conf.Branch)),
			gitconf.RefSpec(fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", conf.MasterBranch, conf.MasterBranch)),
		},
	})

	// Init worktree
	worktree, err := repo.Worktree()
	if err != nil {
		log.Printf("Fail repo.Worktree() %s", err.Error())
	}

	// Checkout master
	err = worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", conf.MasterBranch)),
	})
	if err != nil {
		log.Printf("Fail worktree.Checkout() %s", err.Error())
	}

	// Get master head
	masterCommitRef, err := repo.Head()
	if err != nil {
		log.Printf("Fail repo.Head() %s", err.Error())
	}

	// Get master head commit object
	masterCommit, err := repo.CommitObject(masterCommitRef.Hash())
	if err != nil {
		log.Printf("Fail repo.CommitObject %s", err.Error())
	}
	log.Printf("Master commit: %s", masterCommitRef.Hash())

	// Checkout branch commit
	err = worktree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(conf.Commit),
	})
	if err != nil {
		log.Printf("Fail worktree.Checkout() %s", err.Error())
	}

	// Get branch commit object
	branchCommit, err := repo.CommitObject(plumbing.NewHash(conf.Commit))
	if err != nil {
		log.Printf("Fail branch repo.CommitObject %s", err.Error())
	}
	log.Printf("Branch commit: %s", branchCommit.Hash)

	// Get patch object
	patch, err := masterCommit.Patch(branchCommit)
	if err != nil {
		log.Printf("Fail branchCommit.Patch: %s", err.Error())
	}

	// Iterate through patches
	newFiles := []string{}
	modifiedFiles := []string{}
	deletedFiles := []string{}
	renamedFromFiles := []string{}
	renamedToFiles := []string{}
	filePatches := patch.FilePatches()
	fmt.Printf("Length of patches: %d\n", len(filePatches))
	for _, filePatch := range filePatches {
		fromFile, toFile := filePatch.Files()
		if fromFile == nil {
			fmt.Printf("[+] %s\n", toFile.Path())
			newFiles = append(newFiles, toFile.Path())
		} else if toFile == nil {
			fmt.Printf("[-] %s\n", fromFile.Path())
			deletedFiles = append(deletedFiles, fromFile.Path())
		} else {
			if fromFile.Path() == toFile.Path() {
				fmt.Printf("[~] %s\n", fromFile.Path())
				modifiedFiles = append(modifiedFiles, fromFile.Path())
			} else {
				fmt.Printf("[~] %s to %s", fromFile.Path(), toFile.Path())
				renamedFromFiles = append(renamedFromFiles, fromFile.Path())
				renamedToFiles = append(renamedToFiles, toFile.Path())
			}
		}
	}

}
