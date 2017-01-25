package secret

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/cli"
	"github.com/docker/docker/cli/command"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/system"
	runconfigopts "github.com/docker/docker/runconfig/opts"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

type updateOptions struct {
	name   string
	file   string
	labels opts.ListOpts
}

func newSecretUpdateCommand(dockerCli *command.DockerCli) *cobra.Command {
	updateOpts := updateOptions{
		labels: opts.NewListOpts(opts.ValidateEnv),
	}

	cmd := &cobra.Command{
		Use:   "update [OPTIONS] SECRET",
		Short: "Update a secret from a file or STDIN as content",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updateOpts.name = args[0]
			return runSecretUpdate(dockerCli, updateOpts)
		},
	}
	flags := cmd.Flags()
	flags.VarP(&updateOpts.labels, "label", "l", "Secret labels")
	flags.StringVarP(&updateOpts.file, "file", "f", "", "Read from a file or STDIN ('-')")

	return cmd
}

func runSecretUpdate(dockerCli *command.DockerCli, options updateOptions) error {
	client := dockerCli.Client()
	ctx := context.Background()

	if options.file == "" {
		return fmt.Errorf("Please specify either a file name or STDIN ('-') with --file")
	}

	var in io.Reader = dockerCli.In()
	if options.file != "-" {
		file, err := system.OpenSequential(options.file)
		if err != nil {
			return err
		}
		in = file
		defer file.Close()
	}

	secretData, err := ioutil.ReadAll(in)
	if err != nil {
		return fmt.Errorf("Error reading content from %q: %v", options.file, err)
	}

	spec := swarm.SecretSpec{
		Annotations: swarm.Annotations{
			Name:   options.name,
			Labels: runconfigopts.ConvertKVStringsToMap(options.labels.GetAll()),
		},
		Data: secretData,
	}

	r, err := client.SecretCreate(ctx, spec)
	if err != nil {
		return err
	}

	fmt.Fprintln(dockerCli.Out(), r.ID)
	return nil
}
