package run

import (
	"github.com/iost-official/go-iost/itest"
	"github.com/urfave/cli"
)

// BenchmarkCommand is the command of benchmark
var BenchmarkCommand = cli.Command{
	Name:      "benchmark",
	ShortName: "bench",
	Usage:     "run benchmark test by data",
	Action:    BenchmarkAction,
}

// BenchmarkAction is the action of benchmark
var BenchmarkAction = func(c *cli.Context) error {
	dfile := "benchmark.json"

	r := itest.NewRunner(dfile)
	if err := r.Run(); err != nil {
		return err
	}

	<-r.Done()
	if err := r.Err(); err != nil {
		return err
	}

	return nil
}
