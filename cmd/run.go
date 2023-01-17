package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var imageName string
var execArgs []string
var parallel int
var execTime int
var dataPath string
var dataServer string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Submit and run rhino job",
	// TODO: edit Usage(run build ...)
	Long: "\nSubmit and run rhino job",
	Example: `  rhino run hello:v1.0
  rhino run foo/matmul:v2.1 --np 4 -- arg1 arg2 
  rhino run mpi/testbench -n 32 -t 800 --server 10.0.0.7 --dir /mnt -- --in=/data/file --out=/data/out`,
	RunE: func(cmd *cobra.Command, args []string) error{
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		fmt.Println("print yaml and apply to k8s")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVar(&dataServer, "server", "", "NFS server ip")
	runCmd.Flags().StringVar(&dataPath, "dir", "", "shared directory in NFS server")
	runCmd.MarkFlagsRequiredTogether("server", "dir")
	runCmd.Flags().IntVarP(&parallel, "np", "n", 1, "mpi processes")
	runCmd.Flags().IntVarP(&execTime, "ttl", "t", 600, "estimated execution time(s)")
}
