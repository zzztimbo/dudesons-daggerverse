package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"strings"
)

func (c *Ci) Node(ctx context.Context, testDataSrc *Directory) error {
	var eg errgroup.Group

	eg.Go(func() error {
		refs, err := dag.
			Pipeline("Lazy mode pipeline with oci build").
			Node().
			WithAutoSetup(
				"testdata-myapi",
				testDataSrc.Directory("myapi"),
			).
			Pipeline(
				ctx,
				NodePipelineOpts{
					DryRun: true,
					TTL:    "5m",
					IsOci:  true,
				},
			)

		fmt.Println("image: " + refs)

		return err
	})

	eg.Go(func() error {
		refs, err := dag.
			Pipeline("Explicit mode pipeline with oci build").
			Node().
			WithPipelineID("testdata-myapi").
			WithVersion("20.9.0").
			WithSource(testDataSrc.Directory("myapi")).
			WithNpm().
			Install().
			Test().
			Build().
			OciBuild(ctx, nil, NodeOciBuildOpts{IsTTL: true, TTL: "5m"})

		fmt.Println("image: " + strings.Join(refs, "\n"))

		return err
	})

	eg.Go(func() error {
		_, err := dag.
			Pipeline("Lazy mode pipeline with package build").
			Node().
			WithAutoSetup(
				"testdata-lib",
				testDataSrc.Directory("mylib"),
			).
			Pipeline(
				ctx,
				NodePipelineOpts{
					DryRun:        true,
					PackageDevTag: "beta",
				},
			)

		return err
	})

	eg.Go(func() error {
		_, err := dag.
			Pipeline("Explicit mode pipeline with package build").
			Node().
			WithPipelineID("testdata-mylib").
			WithVersion("20.9.0").
			WithSource(testDataSrc.Directory("mylib")).
			WithNpm().
			Install().
			Test().
			Build().
			Publish(NodePublishOpts{DryRun: true, DevTag: "beta"}).
			Do(ctx)

		return err
	})

	return eg.Wait()
}
