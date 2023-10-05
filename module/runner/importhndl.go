// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

package runner

import (
	"strings"

	"github.com/swaros/contxt/module/ctemplate"
	"github.com/swaros/contxt/module/dirhandle"
	"github.com/swaros/contxt/module/mimiclog"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

type ImportHandler struct {
	imports  []string
	logger   mimiclog.Logger
	dataHndl *tasks.CombinedDh
	template *ctemplate.Template
}

func NewImportHandler(logger mimiclog.Logger, dataHndl *tasks.CombinedDh, template *ctemplate.Template) *ImportHandler {
	return &ImportHandler{
		logger:   logger,
		dataHndl: dataHndl,
		template: template,
	}
}

func (ih *ImportHandler) SetImports(imports []string) {
	ih.imports = imports
}

func (ih *ImportHandler) HandleImports() error {
	if ih.imports == nil {
		return nil
	}
	return ih.handleFileImportsToVars(ih.imports)
}

func (ih *ImportHandler) handleFileImportsToVars(imports []string) error {
	for _, filenameFull := range imports {
		var keyname string
		parts := strings.Split(filenameFull, " ")
		filename := parts[0]
		// loading the file ad parses the template markup inside
		if content, err := ih.template.TryHandleTemplate(filename); err != nil {
			ih.logger.Error("error while loading import", filename)
			return err
		} else {
			ih.logger.Debug("HandleImports: import file loaded", filename)
			if len(parts) > 1 {
				keyname = parts[1]
			}
			var lastErr error
			dirhandle.FileTypeHandler(filename, func(jsonBaseName string) {
				ih.logger.Debug("loading json File as second level variables:", filename)
				if keyname == "" {
					keyname = jsonBaseName
				}
				if err := ih.dataHndl.AddJSON(keyname, content); err != nil {
					ih.logger.Error("error while loading import", filename)
					lastErr = err
				}

			}, func(yamlBaseName string) {
				ih.logger.Debug("loading yaml File: as second level variables", filename)
				if keyname == "" {
					keyname = yamlBaseName
				}
				if err := ih.dataHndl.AddYaml(keyname, content); err != nil {
					ih.logger.Error("error while loading import", filename)
					lastErr = err
				}

			}, func(filenameBase string, ext string) {
				if keyname == "" {
					keyname = filename
				}
				ih.logger.Debug("using content as plain named variable: "+keyname, ext)
				ih.dataHndl.SetPH(keyname, content)

			}, func(path string, err error) {
				ih.logger.Error("file not exists:", err)
				systools.Exit(1)
			})
			// in case of error, return it and stop processing
			if lastErr != nil {
				return lastErr
			}
		}
	}
	return nil
}
