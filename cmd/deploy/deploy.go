package main

import (
	"flag"
	"fmt"
	"github.com/liamg/tml"
	"github.com/sparetimecoders/build-tools/pkg/ci"
	"github.com/sparetimecoders/build-tools/pkg/config"
	"github.com/sparetimecoders/build-tools/pkg/deploy"
	"github.com/sparetimecoders/build-tools/pkg/kubectl"
	ver "github.com/sparetimecoders/build-tools/pkg/version"
	"io"
	"os"
	"time"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	exitFunc = os.Exit
	out io.Writer = os.Stdout
)

func main() {
	if ver.PrintVersionOnly(version, commit, date, out) {
		exitFunc(0)
	} else {
		exitFunc(doDeploy())
	}
}

func doDeploy() int {
	var context, namespace string
	const (
		contextUsage   = "override the context for default environment deployment target"
		namespaceUsage = "override the namespace for default environment deployment target"
	)
	set := flag.NewFlagSet("deploy", flag.ExitOnError)
	set.Usage = func() {
		fmt.Printf("Usage: deploy [options] <environment>\n\nFor example `deploy --context test-cluster --namespace test prod` would deploy to namsepace `test` in the `test-cluster` but assuming to use the `prod` configuration files (if present)\n\nOptions:\n")
		set.PrintDefaults()
	}
	set.StringVar(&context, "context", "", contextUsage)
	set.StringVar(&context, "c", "", contextUsage+" (shorthand)")
	set.StringVar(&namespace, "namespace", "", namespaceUsage)
	set.StringVar(&namespace, "n", "", namespaceUsage+" (shorthand)")
	_ = set.Parse(os.Args[1:])
	if set.NArg() < 1 {
		set.Usage()
	} else {
		environment := set.Args()[0]
		dir, _ := os.Getwd()

		if cfg, err := config.Load(dir, os.Stdout); err != nil {
			fmt.Println(err.Error())
			return -1
		} else {
			if env, err := cfg.CurrentEnvironment(environment); err != nil {
				fmt.Println(err.Error())
			} else {
				if context != "" {
					env.Context = context
				}
				if namespace != "" {
					env.Namespace = namespace
				}
				currentCI := cfg.CurrentCI()
				if !ci.IsValid(currentCI) {
					_, _ = fmt.Println(tml.Sprintf("Commit and/or branch information is <red>missing</red>. Perhaps your not in a Git repository or forgot to set environment variables?"))
					return -2
				}

				tstamp := time.Now().Format(time.RFC3339)
				client := kubectl.New(env, os.Stdout, os.Stderr)
				defer client.Cleanup()
				if err := deploy.Deploy(dir, currentCI.Commit(), currentCI.BuildName(), tstamp, environment, client, os.Stdout, os.Stderr); err != nil {
					fmt.Println(err.Error())
					return -3
				}
			}
		}
	}
	return 0
}
