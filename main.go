package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

func Execute() error {
	flags := struct {
		Default  string
		Password bool
		Select   bool
		Confirm  bool
		Edit     bool
	}{}

	cmd := &cobra.Command{
		Use:          "ask <question>",
		Short:        "ask a question",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := os.Stdin
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				i, err := os.Open("/dev/tty")
				if err != nil {
					return err
				}
				defer i.Close()
				input = i
			}

			var response string
			var prompt survey.Prompt
			if flags.Password {
				prompt = &survey.Password{
					Message: args[0],
				}
			} else if flags.Confirm {
				defaultValue := false
				if cmd.Flags().Changed("default") {
					d, err := strconv.ParseBool(flags.Default)
					if err != nil {
						return fmt.Errorf("invalid default value: %w", err)
					}

					defaultValue = d
				}

				prompt = &survey.Confirm{
					Message: args[0],
					Default: defaultValue,
				}
			} else if flags.Select {
				if isatty.IsTerminal(os.Stdin.Fd()) {
					return fmt.Errorf("stdin is a terminal")
				}

				stdin, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				stdin = bytes.Trim(stdin, "\n")
				stdin = bytes.Trim(stdin, "\r")

				rows := strings.Split(string(stdin), "\n")
				if len(rows) == 0 {
					return fmt.Errorf("no rows")
				}

				prompt = &survey.Select{
					Message: args[0],
					Options: rows,
				}
			} else if flags.Edit {
				var defaultText string
				if !isatty.IsTerminal(os.Stdin.Fd()) {
					stdin, err := io.ReadAll(os.Stdin)
					if err != nil {
						return err
					}
					defaultText = string(stdin)
				}

				prompt = &survey.Editor{
					Message:       args[0],
					Default:       defaultText,
					HideDefault:   true,
					AppendDefault: true,
				}
			} else {
				prompt = &survey.Input{
					Message: args[0],
					Default: flags.Default,
				}
			}

			if err := survey.AskOne(prompt, &response, survey.WithStdio(input, os.Stderr, os.Stderr)); err != nil {
				return err
			}

			fmt.Print(response)
			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.Password, "password", false, "password input")
	cmd.Flags().BoolVar(&flags.Select, "select", false, "select from a list of options")
	cmd.Flags().BoolVar(&flags.Confirm, "confirm", false, "ask for confirmation")
	cmd.Flags().BoolVar(&flags.Edit, "edit", false, "open the user's editor to enter text")
	cmd.Flags().StringVar(&flags.Default, "default", "", "default value")
	cmd.MarkFlagsMutuallyExclusive("password", "default")

	return cmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}
