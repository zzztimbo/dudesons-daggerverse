package main

import (
	"context"
	"golang.org/x/sync/errgroup"
	"strconv"
)

// Return the Node container with the right base image
func (n *Node) WithVersion(
	// The image name to use
	// +optional
	// +default="node"
	image string,
	// The version of the image to use
	version string,
	// Define if the image to use is an alpine or not
	// +optional
	// +default="true"
	isAlpine bool,
) *Node {
	baseImage := image + ":" + version
	if isAlpine {
		baseImage += "-alpine"
	}
	n.Ctr = dag.Container().From(baseImage)

	n.BaseImageRef = baseImage

	return n
}

// Return the Node container with an environment variable to use in your npmrc file
func (n *Node) WithNpmrcTokenEnv(
	// The name of the environment variable where the npmrc token is stored
	name string,
	// The value of the token
	value *Secret,
) *Node {
	n.NpmrcTokenName = name
	n.NpmrcToken = value
	n.Ctr = n.Ctr.WithSecretVariable(name, value)

	return n
}

// Return the Node container with npmrc file
func (n *Node) WithNpmrcTokenFile(
	// The npmrc file to inject in the container
	npmrcFile *Secret,
) *Node {
	n.NpmrcFile = npmrcFile
	n.Ctr = n.Ctr.WithMountedSecret(workdir+"/.npmrc", npmrcFile)

	return n
}

// Return the Node container setup with the right package manager and optionaly cache setup
func (n *Node) WithPackageManager(
	// The package manager to use
	packageManager string,
	// Disable mounting cache volumes.
	// +optional
	disableCache bool,
) *Node {
	switch packageManager {
	case "npm":
		return n.WithNpm(disableCache)
	case "yarn":
		return n.WithYarn(disableCache)
	default:
		return n.WithNpm(disableCache)
	}
}

// Return the Node container with npm setup as an entrypoint and npm cache setup
func (n *Node) WithNpm(
	// Disable mounting cache volumes.
	// +optional
	disableCache bool,
) *Node {
	n.PkgMgr = "npm"
	n.Ctr = n.Ctr.
		WithEntrypoint([]string{"npm"})

	if !disableCache {
		n.Ctr = n.
			Ctr.
			WithMountedCache("/root/.npm", dag.CacheVolume(n.getCacheKey("global-npm-cache")))
	}

	return n
}

// Return the Node container with yarn setup as an entrypoint and yarn cache setup
func (n *Node) WithYarn(
	// Disable mounting cache volumes.
	// +optional
	disableCache bool,
) *Node {
	n.PkgMgr = "yarn"
	n.Ctr = n.Ctr.
		WithEntrypoint([]string{"yarn"})

	if !disableCache {
		n.Ctr = n.
			Ctr.
			WithMountedCache("/usr/local/share/.cache/yarn", dag.CacheVolume(n.getCacheKey("global-yarn-cache")))
	}

	return n
}

// Return the Node container with the source code, 'node_modules' cache set up and workdir set
func (n *Node) WithSource(
	// The source code
	src *Directory,
	// Indicate if the directory is mounted or persisted in the container
	// +optional
	persisted bool,
) *Node {
	if persisted {
		n.Ctr = n.
			Ctr.
			WithDirectory(workdir, src).
			WithWorkdir(workdir)
	} else {
		n.Ctr = n.
			Ctr.
			WithMountedDirectory(workdir, src).
			WithMountedCache(workdir+"/node_modules", dag.CacheVolume(n.getCacheKey("node-modules")))
	}

	n.Ctr = n.Ctr.WithWorkdir(workdir)

	return n
}

// Return a node container with the 'NODE_ENV' set to production
func (n *Node) Production() *Node {
	n.IsProduction = true

	n.Ctr = n.
		Ctr.
		WithEnvVariable("NODE_ENV", "production")
	return n
}

// Execute a command from the package.json
func (n *Node) Run(
	// Command from the package.json to run
	command []string,
) *Node {
	n.Ctr = n.
		Ctr.
		WithExec(append([]string{"run"}, command...))
	return n
}

// Install node modules
func (n *Node) Install() *Node {
	n.Ctr = n.Ctr.WithExec([]string{"install"})
	return n
}

// Execute lint command
func (n *Node) Lint() *Node {
	return n.Run([]string{"lint"})
}

// Execute test command
func (n *Node) Test(
	// Define folder names to mount for testing, these names match 'folder-artifacts'
	// +optional
	folderArtifactNames []string,
	// Define folders to map in the working directory for testing, these folders match 'folder-artifact-names'
	// +optional
	folderArtifacts []*Directory,
	// Define files to mount in the working directory for testing, these names match 'file-artifact-names'
	// +optional
	fileArtifactNames []string,
	// Define file names to map in the working directory for testing, these names match 'file-artifacts'
	// +optional
	fileArtifacts []*File,
	// Define artifact names to mount for testing, these names match 'cache-artifacts'
	// +optional
	cacheArtifactNames []string,
	// Define artifact to map in the working directory for testing, these folders match 'cache-artifact-names'
	// +optional
	cacheArtifacts []string,
) *Node {
	for i, name := range folderArtifactNames {
		n.Ctr = n.
			Ctr.
			WithMountedDirectory(workdir+"/"+name, folderArtifacts[i])
	}

	for i, name := range fileArtifactNames {
		n.Ctr = n.
			Ctr.
			WithMountedFile(workdir+"/"+name, fileArtifacts[i])
	}

	for i, name := range cacheArtifactNames {
		n.Ctr = n.
			Ctr.
			WithMountedCache(workdir+"/"+name, dag.CacheVolume(cacheArtifacts[i]))
	}

	return n.Run([]string{"test"})
}

// Execute test commands in parallel
func (n *Node) ParallelTest(
	ctx context.Context,
	// Define folder names to mount for testing, these names match 'folder-artifacts'
	// +optional
	folderArtifactNames []string,
	// Define folders to map in the working directory for testing, these folders match 'folder-artifact-names'
	// +optional
	folderArtifacts []*Directory,
	// Define files to mount in the working directoryf or testing, these names match 'file-artifact-names'
	// +optional
	fileArtifactNames []string,
	// Define file names to map in the working directory or testing, these names match 'file-artifacts'
	// +optional
	fileArtifacts []*File,
	// Define artifact names to mount for testing or testing, these names match 'cache-artifacts'
	// +optional
	cacheArtifactNames []string,
	// Define artifact to map in the working directory or testing, these folders match 'cache-artifact-names'
	// +optional
	cacheArtifacts []string,
	//Define all command to run
	cmds [][]string,
) error {
	var eg errgroup.Group

	for i, name := range folderArtifactNames {
		n.Ctr = n.
			Ctr.
			WithMountedDirectory(workdir+"/"+name, folderArtifacts[i])
	}

	for i, name := range fileArtifactNames {
		n.Ctr = n.
			Ctr.
			WithMountedFile(workdir+"/"+name, fileArtifacts[i])
	}

	for i, name := range cacheArtifactNames {
		n.Ctr = n.
			Ctr.
			WithMountedCache(workdir+"/"+name, dag.CacheVolume(cacheArtifacts[i]))
	}

	for _, cmd := range cmds {
		eg.Go(func() error {
			_, err := n.Run(cmd).Do(ctx)
			return err
		})
	}

	return eg.Wait()
}

// Execute clean command
func (n *Node) Clean() *Node {
	return n.Run([]string{"clean"})
}

// Todo(think to a method to execute multiple test command, eg: [][]string)

// Execute the build command
func (n *Node) Build() *Node {
	return n.Run([]string{"build"})
}

// Execute the publish which push a package to a registry
func (n *Node) Publish(
	// Define permission on the package in the registry
	// +optional
	access string,
	// Indicate if the package is publishing as development version
	// +optional
	devTag string,
	// Indicate to dry run the publishing
	// +optional
	dryRun bool,
) *Node {
	publishCmd := []string{"publish"}

	if access != "" {
		publishCmd = append(publishCmd, []string{"--access", access}...)
	}

	if devTag != "" {
		publishCmd = append(publishCmd, []string{"--tag", devTag}...)
	}

	if dryRun {
		publishCmd = append(publishCmd, []string{"--dry-run", strconv.FormatBool(dryRun)}...)
	}

	n.Ctr = n.Ctr.WithExec(publishCmd)
	return n
}
