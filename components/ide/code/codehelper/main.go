// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	yaml "gopkg.in/yaml.v2"

	gitpod "github.com/gitpod-io/gitpod/gitpod-protocol"
)

const (
	GITPOD_REPO_ROOT             = "GITPOD_REPO_ROOT"
	GITPOD_WORKSPACE_CONTEXT_URL = "GITPOD_WORKSPACE_CONTEXT_URL"
)

func main() {
	log.SetOutput(io.Discard)
	f, err := os.OpenFile(path.Join(os.TempDir(), ".gitpod-code-helper.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err == nil {
		defer f.Close()
		log.SetOutput(f)
	}
	args := []string{}
	if os.Getenv("SUPERVISOR_DEBUG_ENABLE") == "true" {
		args = append(args, "--inspect", "--log=trace")
	}

	extensions, _ := getExtensions()
	for _, ext := range extensions {
		args = append(args, "--install-extension", ext)
	}
	args = append(args, os.Args...)
	log.Println("run /ide/bin/gitpod-code " + strings.Join(args, " "))
	cmd := exec.Command("/ide/bin/gitpod-code", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("CombinedOutput failed")
	}
	log.Println(">>>>>>>>>>>>>>>>>>>>")
	log.Println(string(output))
}

func getExtensions() (extensions []string, err error) {
	workspaceContextUrl := os.Getenv(GITPOD_WORKSPACE_CONTEXT_URL)
	if workspaceContextUrl != "" && strings.Contains(workspaceContextUrl, "github.com") {
		extensions = append(extensions, "github.vscode-pull-request-github")
	}
	repoRoot := os.Getenv(GITPOD_REPO_ROOT)
	if repoRoot == "" {
		return
	}
	data, err := os.ReadFile(path.Join(repoRoot, ".gitpod.yml"))
	if err != nil {
		return
	}
	var config *gitpod.GitpodConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return
	}
	if config == nil || config.Vscode == nil {
		return
	}
	var wg sync.WaitGroup
	var extensionsMu sync.Mutex
	for _, extIdOrUrl := range config.Vscode.Extensions {
		lowerCaseExtension := strings.ToLower(extIdOrUrl)
		if !isUrl(lowerCaseExtension) {
			extensions = append(extensions, lowerCaseExtension)
		} else {
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				location, err := downloadExtension(url)
				if err != nil {
					return
				}
				extensionsMu.Lock()
				extensions = append(extensions, location)
				extensionsMu.Unlock()
			}(extIdOrUrl)
		}
	}
	wg.Wait()
	return
}

func isUrl(lowerCaseIdOrUrl string) bool {
	isUrl, _ := regexp.MatchString(`http[s]?://`, lowerCaseIdOrUrl)
	return isUrl
}

func downloadExtension(url string) (location string, err error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = errors.New("failed to download extension with status " + http.StatusText(resp.StatusCode))
		return
	}
	out, err := os.CreateTemp("", "vsix*.vsix")
	if err != nil {
		return
	}
	defer out.Close()
	if _, err = io.Copy(out, resp.Body); err != nil {
		return
	}
	location = out.Name()
	return
}
