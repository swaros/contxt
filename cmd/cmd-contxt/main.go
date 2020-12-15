package main

import (
	"github.com/spf13/cobra"
	"github.com/swaros/contxt/context/cmdhandle"
)

var rootCmd = &cobra.Command{
	Use:   "cobra",
	Short: "A generator for Cobra based Applications",
	Long: `Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}



func main() {
	cmdhandle.MainExecute()
}
