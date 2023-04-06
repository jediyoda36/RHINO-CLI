/*
 * Copyright 2023 RHINO Team
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
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

type CreateOptions struct {
	language string
}

func NewCreateCommand() *cobra.Command {
	createOpts := &CreateOptions{}
	createCmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new MPI function/project",
		Long:    "\nCreate a new MPI function/project",
		Example: `  C++ function: rhino create func_name -l cpp`,
		Args:    createOpts.argsCheck,
		RunE:    createOpts.runCreate,
	}
	createCmd.Flags().StringVarP(&createOpts.language, "lang", "l", "cpp", "language template to use")
	return createCmd
}

func (c *CreateOptions) argsCheck(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && len(c.language) == 0 {
		cmd.Help()
		return nil
	} else if len(c.language) == 0 {
		return fmt.Errorf("language cannot be empty")
	} else if len(args) == 0 {
		return fmt.Errorf("function or project name cannot be empty")
	}
	if c.language != "cpp" {
		return fmt.Errorf("only supports cpp in this version")
	}
	return nil
}

func (c *CreateOptions) runCreate(cmd *cobra.Command, args []string) error {
	dirName := args[0]
	if _, err := os.Stat(dirName); err == nil {
		return fmt.Errorf("folder %s already exists", dirName)
	}
	if err := os.Mkdir(dirName, 0700); err != nil {
		return fmt.Errorf("folder %s could not be created", dirName)

	}

	if err := generateTemplate(dirName); err != nil {
		return fmt.Errorf("generate template failed: %s", err.Error())
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
