package cmd

import (
	"fmt"
	"os"
	"io"
	"regexp"
	"github.com/spf13/cobra"
)

var language string
var createCmd = &cobra.Command{
	Use:     "create",
	Short: 	 "Create a new mpi function/project",
	Long: 	 "\nCreate a new mpi function/project",
	Example: `  C function:   rhino create func_name --lang c
  C++ function: rhino create func_name -l cpp
  MPI project:  rhino create proj_name --lang dockerfile`,
  	Args: 	 argsCheck,
  	RunE:    runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&language, "lang", "l", "", "language template to use")
}

func argsCheck(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && len(language) == 0 {
		cmd.Help()
		os.Exit(0)
	} else if len(language) == 0{
		return fmt.Errorf("language cannot be empty")
	} else if len(args) == 0 {
		return fmt.Errorf("function or project name cannot be empty")
	}
	validName := regexp.MustCompile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")
	matchString := validName.MatchString(args[0])
	if !matchString {
		fmt.Println("Error: name can only contain a~z, 0~9 and _")
		os.Exit(0)
	}
	if language != "c" && language != "cpp" && language != "dockerfile" {
		return fmt.Errorf("only support c, cpp and dockerfile")
	}
	return nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	if _, err := os.Stat(name); err == nil {
		fmt.Println("Error: folder", name, "already exists")
		os.Exit(0)
	}
	if err := os.Mkdir(name, 0700); err != nil {
		fmt.Println("Error: folder", name, "could not create")
		os.Exit(0)
	}
	// TODO: download file from github, if fail, delete folder
	// if err := copyFile("./template/func/func.dockerfile", "./test/func.dockerfile"); err != nil {
	// 	fmt.Println("Error:", err.Error())
	// 	os.Exit(0)
	// }
	
	return nil
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	
	_, err = io.Copy(destination, source)
	return err
}
