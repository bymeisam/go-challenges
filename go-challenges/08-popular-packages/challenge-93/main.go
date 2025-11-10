package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	output  string
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "myapp",
		Short: "A simple CLI application",
		Long:  "A longer description of the CLI application",
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	return rootCmd
}

func NewGreetCmd() *cobra.Command {
	greetCmd := &cobra.Command{
		Use:   "greet [name]",
		Short: "Greet someone",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := args[0]
			if verbose {
				fmt.Printf("Greeting %s with verbose mode\n", name)
			}
			fmt.Printf("Hello, %s!\n", name)
		},
	}

	return greetCmd
}

func NewVersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("myapp version 1.0.0")
		},
	}

	return versionCmd
}

func Execute() error {
	rootCmd := NewRootCmd()
	rootCmd.AddCommand(NewGreetCmd())
	rootCmd.AddCommand(NewVersionCmd())
	return rootCmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		os.Exit(1)
	}
}
