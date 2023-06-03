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

func NewCmdInput() *cobra.Command {
	flags := struct {
		Message  string
		Password bool
		Default  string
	}{}

	cmd := &cobra.Command{
		Use: "input",
		RunE: func(cmd *cobra.Command, args []string) error {
			var response string

			var prompt survey.Prompt
			if flags.Password {
				prompt = &survey.Password{
					Message: flags.Message,
				}
			} else {
				prompt = &survey.Input{
					Message: flags.Message,
					Default: flags.Default,
				}
			}

			input := os.Stdin
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				i, err := os.Open("/dev/tty")
				if err != nil {
					return err
				}
				defer i.Close()
				input = i
			}

			if err := survey.AskOne(prompt, &response, survey.WithStdio(input, os.Stderr, os.Stderr)); err != nil {
				return err
			}

			fmt.Print(response)
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.Message, "message", "Input...", "message to display")
	cmd.Flags().BoolVar(&flags.Password, "password", false, "password input")
	cmd.Flags().StringVar(&flags.Default, "default", "", "default value")

	cmd.MarkFlagsMutuallyExclusive("password", "default")

	return cmd
}

func NewCmdConfirm() *cobra.Command {
	flags := struct {
		Message string
		Default string
		Usage   string
	}{}

	cmd := &cobra.Command{
		Use: "confirm",
		RunE: func(cmd *cobra.Command, args []string) error {

			defaultValue := false
			if cmd.Flags().Changed("default") {
				d, err := strconv.ParseBool(flags.Default)
				if err != nil {
					return err
				}

				defaultValue = d
			}

			prompt := &survey.Confirm{
				Message: flags.Message,
				Default: defaultValue,
			}

			input := os.Stdin
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				i, err := os.Open("/dev/tty")
				if err != nil {
					return err
				}
				defer i.Close()
				input = i
			}

			var response bool
			if err := survey.AskOne(prompt, &response, survey.WithStdio(input, os.Stderr, os.Stderr)); err != nil {
				return err
			}

			fmt.Print(response)
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.Message, "message", "Confirm...", "message to display")
	cmd.Flags().StringVar(&flags.Default, "default", "", "default value")
	return cmd
}

func NewCmdSelect() *cobra.Command {
	flags := struct {
		message string
	}{}

	cmd := &cobra.Command{
		Use: "select",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			input, err := os.Open("/dev/tty")
			if err != nil {
				return err
			}
			defer input.Close()

			var response string
			prompt := &survey.Select{
				Message: flags.message,
				Options: rows,
			}

			if err := survey.AskOne(prompt, &response, survey.WithStdio(input, os.Stderr, os.Stderr)); err != nil {
				return err
			}

			fmt.Print(response)

			return nil
		},
	}

	cmd.Flags().StringVar(&flags.message, "message", "Select...", "message to display")
	return cmd
}

func NewCmdEdit() *cobra.Command {
	flags := struct {
		message string
	}{}

	cmd := &cobra.Command{
		Use: "edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			var defaultText string
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				stdin, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				defaultText = string(stdin)
			}

			var response string
			prompt := &survey.Editor{
				Message:       flags.message,
				Default:       defaultText,
				HideDefault:   true,
				AppendDefault: true,
			}

			input := os.Stdin
			if !isatty.IsTerminal(os.Stdin.Fd()) {
				i, err := os.Open("/dev/tty")
				if err != nil {
					return err
				}
				defer i.Close()
				input = i
			}

			if err := survey.AskOne(prompt, &response, survey.WithStdio(input, os.Stderr, os.Stderr)); err != nil {
				return err
			}

			fmt.Print(response)
			return nil
		},
	}

	cmd.Flags().StringVar(&flags.message, "message", "Edit...", "message to display")
	return cmd
}

func Execute() error {
	cmd := cobra.Command{
		Use:          "survey",
		SilenceUsage: true,
	}

	cmd.AddCommand(NewCmdInput())
	cmd.AddCommand(NewCmdConfirm())
	cmd.AddCommand(NewCmdSelect())
	cmd.AddCommand(NewCmdEdit())

	return cmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}