package appcli

import (
	"github.com/pkg/errors"
	"github.com/urfave/cli"

	"gitlab.com/rliebz/tusk/config"
	"gitlab.com/rliebz/tusk/task"
)

// copyFlags copies all command flags from one cli.App to another.
func copyFlags(target *cli.App, source *cli.App) {
	for i, targetCommand := range target.Commands {
		for _, sourceCommand := range source.Commands {
			if targetCommand.Name == sourceCommand.Name {
				target.Commands[i].Flags = sourceCommand.Flags
			}
		}
	}
}

// addGlobalFlagsUsed adds the top-level flags to tasks where interpolation is used.
func addGlobalFlagsUsed(cmd *cli.Command, t *task.Task, cfg *config.Config) error {

	dependencies, err := cfg.FindAllOptions(t)
	if err != nil {
		return err
	}

	for _, opt := range dependencies {

		if opt.Private {
			continue
		}

		// TODO: Disallow multiple differing flag definitions

		if err := addFlag(cmd, opt); err != nil {
			return errors.Wrapf(
				err,
				"could not add flag `%s` to command `%s`",
				opt.Name,
				t.Name,
			)
		}

	}

	return nil
}

func addFlag(command *cli.Command, opt *task.Option) error {
	flag, err := task.CreateCLIFlag(opt)
	if err != nil {
		return err
	}

	for _, flag := range command.Flags {
		if opt.Name == flag.GetName() {
			return nil
		}
	}

	command.Flags = append(command.Flags, flag)

	return nil
}
