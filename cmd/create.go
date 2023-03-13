package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/OpenRHINO/RHINO-CLI/generate"
	"github.com/spf13/cobra"
)

var language string
var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a new MPI function/project",
	Long:    "\nCreate a new MPI function/project",
	Example: `  C++ function: rhino create func_name -l cpp`,
	Args:    argsCheck,
	RunE:    runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&language, "lang", "l", "cpp", "language template to use")
}

func argsCheck(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && len(language) == 0 {
		cmd.Help()
		os.Exit(0)
	} else if len(language) == 0 {
		return fmt.Errorf("language cannot be empty")
	} else if len(args) == 0 {
		return fmt.Errorf("function or project name cannot be empty")
	}
	if language != "cpp" {
		return fmt.Errorf("only supports cpp in this version")
	}
	return nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	dirName := args[0]
	if _, err := os.Stat(dirName); err == nil {
		fmt.Println("Error: folder", dirName, "already exists")
		os.Exit(0)
	}
	if err := os.Mkdir(dirName, 0700); err != nil {
		fmt.Println("Error: folder", dirName, "could not be created")
		os.Exit(0)
	}

	if err := generateTemplate(dirName); err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(0)
	}

	return nil
}

func generateTemplate(dstDir string) error {
	zr, err := zip.NewReader(bytes.NewReader(generate.TemplatesZip), int64(len(generate.TemplatesZip)))
	if err != nil {
		return err
	}

	for _, file := range zr.File {
		path := filepath.Join(dstDir, file.Name)
		// 如果是目录，则创建目录，并跳过当前循环，继续处理下一个
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return err
			}
			continue
		}

		// 解压文件到目标文件夹
		fr, err := file.Open()
		if err != nil {
			return err
		}
		fw, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		_, err = io.Copy(fw, fr)
		if err != nil {
			return err
		}
		fw.Close()
		fr.Close()
	}

	return nil
}
