// Copyright 2025 The PipeCD Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	sdk "github.com/pipe-cd/piped-plugin-sdk-go"
)

type plugin struct{}

var _ sdk.StagePlugin[pluginConfig, sdk.ConfigNone, sdk.ConfigNone] = (*plugin)(nil)

type pluginConfig struct {
	AnsiblePath string `json:"ansiblePath"`
	Inventory   string `json:"inventory"`
	Vault       string `json:"vault"`
}

type ansiblePlaybookStageOptions struct {
	Playbook    string            `json:"playbook"`
	Inventory   string            `json:"inventory"`
	ExtraVars   map[string]string `json:"extraVars"`
	Tags        []string          `json:"tags"`
	SkipTags    []string          `json:"skipTags"`
	Limit       string            `json:"limit"`
	Verbosity   int               `json:"verbosity"`
	CheckMode   bool              `json:"checkMode"`
	DiffMode    bool              `json:"diffMode"`
	Vault       string            `json:"vault"`
	PrivateKey  string            `json:"privateKey"`
	RemoteUser  string            `json:"remoteUser"`
	BecomeUser  string            `json:"becomeUser"`
	Timeout     int               `json:"timeout"`
}

func (p *ansiblePlaybookStageOptions) validate() error {
	if p.Playbook == "" {
		return fmt.Errorf("playbook is required")
	}
	if p.Timeout <= 0 {
		p.Timeout = 600
	}
	return nil
}

const (
	stageAnsiblePlaybook = "ANSIBLE_PLAYBOOK"
)

func (p *plugin) FetchDefinedStages() []string {
	return []string{
		stageAnsiblePlaybook,
	}
}

func (p *plugin) BuildPipelineSyncStages(ctx context.Context, _ *pluginConfig, input *sdk.BuildPipelineSyncStagesInput) (*sdk.BuildPipelineSyncStagesResponse, error) {
	return &sdk.BuildPipelineSyncStagesResponse{
		Stages: buildPipelineSyncStages(input.Request),
	}, nil
}

func (p *plugin) ExecuteStage(ctx context.Context, cfg *pluginConfig, _ []*sdk.DeployTarget[sdk.ConfigNone], input *sdk.ExecuteStageInput[sdk.ConfigNone]) (*sdk.ExecuteStageResponse, error) {
	switch input.Request.StageName {
	case stageAnsiblePlaybook:
		return executeAnsiblePlaybook(ctx, cfg, input)
	}
	return nil, fmt.Errorf("unsupported stage: %s", input.Request.StageName)
}

func buildPipelineSyncStages(req sdk.BuildPipelineSyncStagesRequest) []sdk.PipelineStage {
	stages := make([]sdk.PipelineStage, 0, len(req.Stages))
	for _, rs := range req.Stages {
		stage := sdk.PipelineStage{
			Index:              rs.Index,
			Name:               rs.Name,
			Rollback:           false,
			Metadata:           map[string]string{},
			AvailableOperation: sdk.ManualOperationNone,
		}
		stages = append(stages, stage)
	}
	return stages
}

func executeAnsiblePlaybook(ctx context.Context, cfg *pluginConfig, input *sdk.ExecuteStageInput[sdk.ConfigNone]) (*sdk.ExecuteStageResponse, error) {
	lp := input.Client.LogPersister()
	var stageOpts ansiblePlaybookStageOptions
	if err := json.Unmarshal(input.Request.StageConfig, &stageOpts); err != nil {
		lp.Errorf("Failed to unmarshal the stage config (%v)", err)
		return nil, fmt.Errorf("failed to unmarshal the stage config (%v)", err)
	}
	if err := stageOpts.validate(); err != nil {
		lp.Errorf("Invalid stage options: %v", err)
		return nil, fmt.Errorf("invalid stage options: %v", err)
	}

	ansiblePath := cfg.AnsiblePath
	if ansiblePath == "" {
		ansiblePath = "ansible-playbook"
	}

	playbookPath := filepath.Join(input.Request.TargetDeploymentSource.ApplicationDirectory, stageOpts.Playbook)
	if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
		lp.Errorf("Playbook file does not exist: %s", playbookPath)
		return nil, fmt.Errorf("playbook file does not exist: %s", playbookPath)
	}

	args := []string{playbookPath}

	inventory := stageOpts.Inventory
	if inventory == "" {
		inventory = cfg.Inventory
	}
	if inventory != "" {
		inventoryPath := filepath.Join(input.Request.TargetDeploymentSource.ApplicationDirectory, inventory)
		args = append(args, "-i", inventoryPath)
	}

	if len(stageOpts.ExtraVars) > 0 {
		extraVars := make([]string, 0, len(stageOpts.ExtraVars))
		for k, v := range stageOpts.ExtraVars {
			extraVars = append(extraVars, fmt.Sprintf("%s=%s", k, v))
		}
		args = append(args, "--extra-vars", strings.Join(extraVars, " "))
	}

	if len(stageOpts.Tags) > 0 {
		args = append(args, "--tags", strings.Join(stageOpts.Tags, ","))
	}

	if len(stageOpts.SkipTags) > 0 {
		args = append(args, "--skip-tags", strings.Join(stageOpts.SkipTags, ","))
	}

	if stageOpts.Limit != "" {
		args = append(args, "--limit", stageOpts.Limit)
	}

	if stageOpts.Verbosity > 0 {
		verbosity := strings.Repeat("v", stageOpts.Verbosity)
		args = append(args, fmt.Sprintf("-%s", verbosity))
	}

	if stageOpts.CheckMode {
		args = append(args, "--check")
	}

	if stageOpts.DiffMode {
		args = append(args, "--diff")
	}

	vault := stageOpts.Vault
	if vault == "" {
		vault = cfg.Vault
	}
	if vault != "" {
		vaultPath := filepath.Join(input.Request.TargetDeploymentSource.ApplicationDirectory, vault)
		args = append(args, "--vault-password-file", vaultPath)
	}

	if stageOpts.PrivateKey != "" {
		keyPath := filepath.Join(input.Request.TargetDeploymentSource.ApplicationDirectory, stageOpts.PrivateKey)
		args = append(args, "--private-key", keyPath)
	}

	if stageOpts.RemoteUser != "" {
		args = append(args, "--user", stageOpts.RemoteUser)
	}

	if stageOpts.BecomeUser != "" {
		args = append(args, "--become", "--become-user", stageOpts.BecomeUser)
	}

	lp.Infof("Executing ansible-playbook command: %s %s", ansiblePath, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, ansiblePath, args...)
	cmd.Dir = input.Request.TargetDeploymentSource.ApplicationDirectory
	cmd.Stdout = lp
	cmd.Stderr = lp

	if err := cmd.Run(); err != nil {
		lp.Errorf("Failed to execute ansible-playbook: %v", err)
		return &sdk.ExecuteStageResponse{Status: sdk.StageStatusFailure}, fmt.Errorf("failed to execute ansible-playbook: %v", err)
	}

	lp.Infof("Ansible playbook executed successfully")
	return &sdk.ExecuteStageResponse{Status: sdk.StageStatusSuccess}, nil
}