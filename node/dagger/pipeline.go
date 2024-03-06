package main

import (
	"context"
	"strings"
)

// Execute the whole pipeline in general used with the function 'with-auto-setup'
func (n *Node) Pipeline(
	ctx context.Context,
	// Define hooks to execute before all
	// +optional
	preHooks [][]string,
	// Define hooks to execute after tests and before build
	// +optional
	postHooks [][]string,
	// Indicate if the artifact is an oci build or not
	// +optional
	isOci bool,
	// Indicate to dry run the publishing
	// +optional
	// +default="false"
	dryRun bool,
	// Define permission on the package in the registry
	// +optional
	// +default="true"
	packageAccess string,
	// Indicate if the package is publishing as development version
	// +optional
	packageDevTag string,
	// Define folder names to mount for testing, these names match 'folder-artifacts'
	// +optional
	testFolderArtifactNames []string,
	// Define folders to map in the working directory for testing, these folders match 'test-folder-artifact-names'
	// +optional
	testFolderArtifacts []*Directory,
	// Define files to mount in the working directoryf or testing, these names match 'test-file-artifact-names'
	// +optional
	testFileArtifactNames []string,
	// Define file names to map in the working directory or testing, these names match 'test-file-artifacts'
	// +optional
	testFileArtifacts []*File,
	// Define artifact names to mount for testing or testing, these names match 'test-cache-artifacts'
	// +optional
	testCacheArtifactNames []string,
	// Define artifact to map in the working directory or testing, these folders match 'test-cache-artifact-names'
	// +optional
	testCacheArtifacts []string,
	// Define folder names to map in the working directory, these names match 'folder-artifacts'
	// +optional
	ociFolderArtifactNames []string,
	// Define folders to map in the working directory, these folders match 'folder-artifact-names'
	// +optional
	ociFolderArtifacts []*Directory,
	// Define files to mount in the working directory, these names match 'file-artifact-names'
	// +optional
	ociFileArtifactNames []string,
	// Define file names to map in the working directory, these names match 'file-artifacts'
	// +optional
	ociFileArtifacts []*File,
	// Define registries where to push the image
	// +optional
	ociRegistries []string,
	// Define the ttl registry to use
	// +optional
	// +default="ttl.sh"
	ttlRegistry string,
	// Define the ttl in the ttl registry
	// +optional
	// +default="60m"
	ttl string,
) (string, error) {
	pipeline := n.Install()

	for _, hook := range preHooks {
		pipeline = pipeline.Run(hook)
	}

	if n.DetectLint {
		pipeline = pipeline.Lint()
	}

	if n.DetectTest {
		pipeline = pipeline.Test(
			testFolderArtifactNames,
			testFolderArtifacts,
			testFileArtifactNames,
			testFileArtifacts,
			testCacheArtifactNames,
			testCacheArtifacts,
		)
	}

	for _, hook := range postHooks {
		pipeline = pipeline.Run(hook)
	}

	pipeline = pipeline.Build()

	if n.DetectPackage {
		return pipeline.
			Publish(packageAccess, packageDevTag, dryRun).
			Do(ctx)
	}

	if n.DetectOci || isOci {
		refs, err := pipeline.
			OciBuild(
				ctx,
				ociFolderArtifactNames,
				ociFolderArtifacts,
				ociFileArtifactNames,
				ociFileArtifacts,
				nil,
				nil,
				ociRegistries,
				dryRun,
				ttlRegistry,
				ttl,
			)

		return strings.Join(refs, "\n"), err
	}
	return pipeline.Do(ctx)
}
