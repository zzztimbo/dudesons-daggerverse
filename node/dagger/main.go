// A generated module for Node functions
//
// This module has been generated via dagger init and serves as a reference to
// basic module structure as you get started with Dagger.
//
// Two functions have been pre-created. You can modify, delete, or add to them,
// as needed. They demonstrate usage of arguments and return types using simple
// echo and grep commands. The functions can be called from the dagger CLI or
// from one of the SDKs.
//
// The first line in this comment block is a short description line and the
// rest is a long description with more detail on the module's purpose or usage,
// if appropriate. All modules should have a short description.

package main

import (
	"context"
)

const (
	workdir = "/opt/app"
)

type Node struct {
	// +private
	PipelineID string
	// +private
	Ctr *Container
	// +private
	Name string
	// +private
	Version string
	// +private
	DetectTest bool
	// +private
	DetectPackage bool
	// +private
	DetectLint bool
	// +private
	DetectOci bool
	// +private
	PkgMgr string
	// +private
	Platform Platform
	// +private
	IsProduction bool
	// +private
	SystemSetupCmds [][]string
	// +private
	BaseImageRef string
	// +private
	NpmrcTokenName string
	// +private
	NpmrcToken *Secret
	// +private
	NpmrcFile *Secret
	// +private
	DistName string
}

// Define the pipeline id to use
func (n *Node) WithPipelineId(
	// The name to give to the pipeline
	pipelineID string,
) *Node {
	n.PipelineID = pipelineID

	return n
}

// Setup system component like installing packages
func (n *Node) SetupSystem(
	// Indicate attempted system package to install
	// +optional
	systemSetupCmds [][]string,
) *Node {
	n.SystemSetupCmds = append(n.SystemSetupCmds, systemSetupCmds...)

	for _, i := range n.SystemSetupCmds {
		n.Ctr = n.Ctr.WithExec(i)
	}

	return n
}

// Execute all commands
func (n *Node) Do(ctx context.Context) (string, error) {
	return n.Ctr.Stdout(ctx)
}
