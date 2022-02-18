// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"

	gitpod "github.com/gitpod-io/gitpod/gitpod-protocol"
	supervisor "github.com/gitpod-io/gitpod/supervisor/api"
)

func main() {
	// Wait until content ready
	contentStatus, wsInfo, err := resolveWorkspaceInfo(context.Background())
	if err != nil || wsInfo == nil || contentStatus == nil || !contentStatus.Available {
		return
	}
	output := "--start-server"
	if strings.Contains(wsInfo.GetWorkspaceContextUrl(), "github.com") {
		output += " --install-builtin-extension github.vscode-pull-request-github"
	}
	uniqMap := map[string]struct{}{}
	extensions, _ := getExtensions(wsInfo.GetCheckoutLocation())
	for _, ext := range extensions {
		if _, ok := uniqMap[ext]; ok {
			continue
		}
		uniqMap[ext] = struct{}{}
		output += " --install-extension " + ext
	}
	fmt.Print(output)
}

func getExtensions(repoRoot string) (extensions []string, err error) {
	if repoRoot == "" {
		return
	}
	data, err := os.ReadFile(filepath.Join(repoRoot, ".gitpod.yml"))
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
	for _, ext := range config.Vscode.Extensions {
		lowerCaseExtension := strings.ToLower(ext)
		if isUrl(lowerCaseExtension) {
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
			}(ext)
		} else {
			path := filepath.Join(repoRoot, ext)
			if isVsixFileExists(path) {
				extensions = append(extensions, path)
			} else {
				extensions = append(extensions, lowerCaseExtension)
			}
		}
	}
	wg.Wait()
	return
}

func isUrl(lowerCaseIdOrUrl string) bool {
	isUrl, _ := regexp.MatchString(`http[s]?://`, lowerCaseIdOrUrl)
	return isUrl
}

func isVsixFileExists(path string) bool {
	if !strings.HasSuffix(path, ".vsix") {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
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

func resolveWorkspaceInfo(ctx context.Context) (contentStatus *supervisor.ContentStatusResponse, wsInfo *supervisor.WorkspaceInfoResponse, err error) {
	supervisorAddr := os.Getenv("SUPERVISOR_ADDR")
	if supervisorAddr == "" {
		supervisorAddr = "localhost:22999"
	}
	supervisorConn, err := grpc.Dial(supervisorAddr, grpc.WithInsecure())
	if err != nil {
		return
	}
	defer supervisorConn.Close()
	if wsInfo, err = supervisor.NewInfoServiceClient(supervisorConn).WorkspaceInfo(ctx, &supervisor.WorkspaceInfoRequest{}); err != nil {
		return
	}
	contentStatus, err = supervisor.NewStatusServiceClient(supervisorConn).ContentStatus(ctx, &supervisor.ContentStatusRequest{Wait: true})
	return
}
