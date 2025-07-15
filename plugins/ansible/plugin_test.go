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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlugin_FetchDefinedStages(t *testing.T) {
	p := &plugin{}
	stages := p.FetchDefinedStages()
	
	assert.Len(t, stages, 1)
	assert.Contains(t, stages, "ANSIBLE_PLAYBOOK")
}

func TestAnsiblePlaybookStageOptions_validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    ansiblePlaybookStageOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: ansiblePlaybookStageOptions{
				Playbook: "playbook.yml",
				Timeout:  300,
			},
			wantErr: false,
		},
		{
			name: "missing playbook",
			opts: ansiblePlaybookStageOptions{
				Timeout: 300,
			},
			wantErr: true,
		},
		{
			name: "sets default timeout",
			opts: ansiblePlaybookStageOptions{
				Playbook: "playbook.yml",
				Timeout:  0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.opts.Timeout == 0 {
					assert.Equal(t, 600, tt.opts.Timeout)
				}
			}
		})
	}
}

func TestStageOptionsUnmarshal(t *testing.T) {
	jsonData := `{
		"playbook": "deploy.yml",
		"inventory": "hosts",
		"extraVars": {
			"env": "production",
			"version": "1.0.0"
		},
		"tags": ["deploy", "configure"],
		"verbosity": 2,
		"checkMode": true
	}`

	var opts ansiblePlaybookStageOptions
	err := json.Unmarshal([]byte(jsonData), &opts)
	assert.NoError(t, err)
	
	assert.Equal(t, "deploy.yml", opts.Playbook)
	assert.Equal(t, "hosts", opts.Inventory)
	assert.Equal(t, map[string]string{"env": "production", "version": "1.0.0"}, opts.ExtraVars)
	assert.Equal(t, []string{"deploy", "configure"}, opts.Tags)
	assert.Equal(t, 2, opts.Verbosity)
	assert.True(t, opts.CheckMode)
}