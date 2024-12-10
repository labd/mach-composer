package runner

import (
	"context"
	"fmt"
	"github.com/mach-composer/mach-composer-cli/internal/batcher"
	"github.com/mach-composer/mach-composer-cli/internal/cli"
	"github.com/mach-composer/mach-composer-cli/internal/graph"
	"github.com/mach-composer/mach-composer-cli/internal/hash"
	"github.com/mach-composer/mach-composer-cli/internal/terraform"
	"github.com/mach-composer/mach-composer-cli/internal/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/semaphore"
	"sort"
	"sync"
	"time"
)

type (
	//executorFunc is a function that executes an arbitrary command on a node
	executorFunc func(ctx context.Context, node graph.Node) error
)

// GraphRunner will run a set of commands on a graph of nodes. Untainted nodes (no changes) will be skipped.
// The nodes are batched based on a batching function, and all nodes in the same batch will be run in parallel.
type GraphRunner struct {
	workers int
	batch   batcher.BatchFunc
	hash    hash.Handler
}

func NewGraphRunner(batcher batcher.BatchFunc, hashHandler hash.Handler, workers int) *GraphRunner {
	return &GraphRunner{
		workers: workers,
		batch:   batcher,
		hash:    hashHandler,
	}
}

func (gr *GraphRunner) run(ctx context.Context, g *graph.Graph, f executorFunc, ignoreChangeDetection bool) error {
	if err := taintGraph(ctx, g, gr.hash); err != nil {
		return err
	}

	batches := gr.batch(g)

	keys := maps.Keys(batches)
	sort.Ints(keys)
	for i, k := range keys[1:] {
		log.Info().Msgf("Running batch %d with %d nodes", i, len(batches[k]))

		errChan := make(chan error, len(batches[k]))
		var outputs []*cli.BufferedWriter

		wg := &sync.WaitGroup{}
		sem := semaphore.NewWeighted(int64(gr.workers))

		// Start a ticker to show we are still running
		ticker := make(chan struct{})
		go func() {
			for {
				select {
				case <-ticker:
					return
				default:
					log.Info().Msgf("Waiting for batch %d to complete", i)
					time.Sleep(1 * time.Second)
				}
			}
		}()

		for _, n := range batches[k] {
			if n.Tainted() == false && ignoreChangeDetection == false {
				log.Info().Msgf("Skipping %s because it has no changes", n.Identifier())
				continue
			}

			if n.Targeted() == false {
				log.Info().Msgf("Skipping %s because it is not targeted", n.Identifier())
				continue
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			wg.Add(1)

			go func(ctx context.Context, n graph.Node) {
				defer wg.Done()
				defer sem.Release(1)

				w := cli.LogWriterFromContext(ctx)
				bw := cli.NewBufferedWriter(w)
				l := log.Output(bw).With().Str("identifier", n.Identifier()).Logger()
				ctx = l.WithContext(ctx)

				outputs = append(outputs, bw)

				defer func() {
					if cli.GithubCIFromContext(ctx) {
						log.Ctx(ctx).Info().Msg("::endgroup::")
					}
				}()

				if cli.GithubCIFromContext(ctx) {
					log.Ctx(ctx).Info().Msgf("::group::{%s}", n.Identifier())
				}

				err := f(ctx, n)
				if err != nil {
					errChan <- err
					return
				}
			}(ctx, n)
		}
		wg.Wait()
		close(ticker)
		close(errChan)

		for _, output := range outputs {
			if err := output.Flush(); err != nil {
				return err
			}
		}

		if len(errChan) > 0 {
			var errors []error
			for err := range errChan {
				errors = append(errors, err)
			}

			return cli.NewGroupedError(fmt.Sprintf("batch run %d failed (%d errors)", i, len(errors)), errors)
		}

		log.Info().Msgf("Finished batch %d", i)
	}

	log.Info().Msgf("Finished all batches")

	return nil
}

func (gr *GraphRunner) TerraformApply(ctx context.Context, dg *graph.Graph, opts *ApplyOptions) error {
	if err := gr.run(ctx, dg, func(ctx context.Context, n graph.Node) error {
		if !terraformIsInitialized(ctx, n.Path()) || opts.ForceInit {
			log.Ctx(ctx).Info().Msgf("Running terraform init for %s", n.Path())
			out, err := terraform.Init(ctx, n.Path())
			if err != nil {
				return err
			}
			log.Ctx(ctx).Info().Msg(out)
		} else {
			log.Ctx(ctx).Info().Msgf("Skipping terraform init for %s", n.Path())
		}

		var aOpts []terraform.ApplyOption
		if opts.Destroy {
			aOpts = append(aOpts, terraform.ApplyWithDestroy())
		}
		if opts.AutoApprove {
			aOpts = append(aOpts, terraform.ApplyWithAutoApprove())
		}
		if cli.OutputFromContext(ctx) == cli.OutputTypeJSON {
			aOpts = append(aOpts, terraform.ApplyWithJson())
		}

		out, err := terraform.Apply(ctx, n.Path(), aOpts...)
		if err != nil {
			err = fmt.Errorf("failed to apply %s: %w", n.Identifier(), err)
		}

		if cli.OutputFromContext(ctx) == cli.OutputTypeJSON {
			logLines, err := cli.ParseTerraformJsonOutput(out)
			if err != nil {
				return err
			}
			for _, logLine := range logLines {
				level, err := zerolog.ParseLevel(logLine.Level)
				if err != nil {
					level = zerolog.InfoLevel
				}
				log.Ctx(ctx).WithLevel(level).Fields(logLine.Remainder).Msg(logLine.Message)
			}
		} else {
			log.Ctx(ctx).Info().Msg(out)
		}

		log.Ctx(ctx).Debug().Msgf("Storing new hash for %s", n.Path())
		if err := gr.hash.Store(ctx, n); err != nil {
			log.Ctx(ctx).Warn().Err(err).Msgf("Failed to store hash for %s", n.Identifier())
		}

		return err

	}, opts.IgnoreChangeDetection); err != nil {
		return err
	}

	return nil
}

func (gr *GraphRunner) TerraformValidate(ctx context.Context, dg *graph.Graph) error {
	return gr.run(ctx, dg, func(ctx context.Context, n graph.Node) error {
		log.Ctx(ctx).Info().Msgf("Running terraform init without backend for %s", n.Path())
		out, err := terraform.Init(ctx, n.Path(), terraform.InitWithDisableBackend())
		if err != nil {
			return err
		}
		log.Ctx(ctx).Info().Msg(out)

		log.Ctx(ctx).Info().Msgf("Running terraform validate for %s", n.Path())

		var vOpts []terraform.ValidateOption

		out, err = terraform.Validate(ctx, n.Path(), vOpts...)
		if err != nil {
			err = fmt.Errorf("failed to validate %s: %w", n.Identifier(), err)
		}
		log.Ctx(ctx).Info().Msg(out)

		return err
	}, true)
}

func (gr *GraphRunner) TerraformPlan(ctx context.Context, dg *graph.Graph, opts *PlanOptions) error {
	if err := gr.run(ctx, dg, func(ctx context.Context, n graph.Node) error {
		if !terraformIsInitialized(ctx, n.Path()) || opts.ForceInit {
			log.Ctx(ctx).Info().Msgf("Running terraform init for %s", n.Path())
			out, err := terraform.Init(ctx, n.Path())
			if err != nil {
				return err
			}
			log.Ctx(ctx).Info().Msg(out)
		} else {
			log.Ctx(ctx).Info().Msgf("Skipping terraform init for %s", n.Path())
		}

		canPlan, err := terraformCanPlan(ctx, n)
		if err != nil {
			return err
		}

		if !canPlan {
			log.Ctx(ctx).Info().Msgf("Skipping planning %s because it has missing outputs", n.Path())
			return err
		}

		var pOpts []terraform.PlanOption
		if !opts.Lock {
			pOpts = append(pOpts, terraform.PlanWithNoLock())
		}
		if cli.OutputFromContext(ctx) == cli.OutputTypeJSON {
			pOpts = append(pOpts, terraform.PlanWithJson())
		}

		out, err := terraform.Plan(ctx, n.Path(), pOpts...)
		if err != nil {
			err = fmt.Errorf("failed to plan %s: %w", n.Identifier(), err)
		}

		if cli.OutputFromContext(ctx) == cli.OutputTypeJSON {
			logLines, err := cli.ParseTerraformJsonOutput(out)
			if err != nil {
				return err
			}
			for _, logLine := range logLines {
				level, err := zerolog.ParseLevel(logLine.Level)
				if err != nil {
					level = zerolog.InfoLevel
				}
				log.Ctx(ctx).WithLevel(level).Fields(logLine.Remainder).Msg(logLine.Message)
			}
		} else {
			log.Ctx(ctx).Info().Msg(out)
		}

		return err
	}, opts.IgnoreChangeDetection); err != nil {
		return err
	}

	return nil
}

func (gr *GraphRunner) TerraformProxy(ctx context.Context, dg *graph.Graph, opts *ProxyOptions) error {
	if err := gr.run(ctx, dg, func(ctx context.Context, n graph.Node) error {
		if !terraformIsInitialized(ctx, n.Path()) {
			return fmt.Errorf("terraform is not initialized for %s. Please run init beforehand", n.Path())
		}

		out, err := utils.RunTerraform(ctx, n.Path(), opts.Command...)
		if err != nil {
			err = fmt.Errorf("failed to proxy %s: %w", n.Identifier(), err)
		}
		log.Ctx(ctx).Info().Msg(out)
		return err
	}, opts.IgnoreChangeDetection); err != nil {
		return err
	}

	return nil
}

func (gr *GraphRunner) TerraformShow(ctx context.Context, dg *graph.Graph, opts *ShowPlanOptions) error {
	if err := gr.run(ctx, dg, func(ctx context.Context, n graph.Node) error {
		if !terraformIsInitialized(ctx, n.Path()) || opts.ForceInit {
			log.Ctx(ctx).Info().Msgf("Running terraform init for %s", n.Path())
			out, err := terraform.Init(ctx, n.Path())
			if err != nil {
				return err
			}
			log.Ctx(ctx).Info().Msg(out)
		} else {
			log.Ctx(ctx).Info().Msgf("Skipping terraform init for %s", n.Path())
		}

		var sOpts []terraform.ShowOption
		if opts.NoColor {
			sOpts = append(sOpts, terraform.ShowWithNoColor())
		}
		if cli.OutputFromContext(ctx) == cli.OutputTypeJSON {
			sOpts = append(sOpts, terraform.ShowWithJson())
		}

		out, err := terraform.Show(ctx, n.Path(), sOpts...)
		if err != nil {
			err = fmt.Errorf("failed to show %s: %w", n.Identifier(), err)
		}

		if cli.OutputFromContext(ctx) == cli.OutputTypeJSON {
			logLines, err := cli.ParseTerraformJsonOutput(out)
			if err != nil {
				return err
			}
			for _, logLine := range logLines {
				level, err := zerolog.ParseLevel(logLine.Level)
				if err != nil {
					level = zerolog.InfoLevel
				}
				log.Ctx(ctx).WithLevel(level).Fields(logLine.Remainder).Msg(logLine.Message)
			}
		} else {
			log.Ctx(ctx).Info().Msg(out)
		}
		return err
	}, opts.IgnoreChangeDetection); err != nil {
		return err
	}

	return nil
}

func (gr *GraphRunner) TerraformInit(ctx context.Context, dg *graph.Graph) error {
	if err := gr.run(ctx, dg, func(ctx context.Context, n graph.Node) error {
		out, err := terraform.Init(ctx, n.Path())
		if err != nil {
			err = fmt.Errorf("failed to init %s: %w", n.Identifier(), err)
		}
		log.Ctx(ctx).Info().Msg(out)
		return err
	}, true); err != nil {
		return err
	}

	return nil
}
