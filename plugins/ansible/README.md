# Ansible Plugin


## Overview

Ansible plugin supports deployment execution using Ansible playbooks for PipeCD.

> [!CAUTION] 
> Currently, this is in development status.

The Ansible plugin allows you to execute Ansible playbooks as part of your deployment pipeline, providing infrastructure automation and configuration management capabilities within PipeCD.

### Quick sync

Quick sync executes the specified Ansible playbook to deploy the application.

It will be planned in one of the following cases:
- no pipeline was specified in the application configuration file
- `pipeline` was specified but the deployment strategy is set to quick sync

For example, the application configuration below is missing the pipeline field. This means any deployment will trigger a quick sync deployment.

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins: 
    ansible:
      kind: Playbook
      playbook:
        path: playbooks/deploy.yml
        inventory: inventory/hosts
        verbosity: 1
```

### Pipeline sync

You can configure the pipeline to enable structured deployment with multiple stages.

These are the provided stages for Ansible plugin you can use to build your pipeline:

- `ANSIBLE_SYNC`
  - execute the specified Ansible playbook to deploy the application

## Plugin Configuration

### Piped Config

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Piped
spec:
  plugins:
  - name: ansible
    url: file:///path/to/.piped/plugins/ansible # or remoteUrl(TBD)
    config:
      ansiblePath: /usr/local/bin/ansible-playbook
      inventory: inventory/default
      vault: vault/password
    deployTargets:
      - name: production
        config:
          ansiblePath: /usr/bin/ansible-playbook
          inventory: inventory/production
          vault: vault/prod-password
      - name: staging
        config:
          ansiblePath: /usr/bin/ansible-playbook
          inventory: inventory/staging
```

| Field | Type | Description | Required |
|-|-|-|-|
| ansiblePath | string | Path to ansible-playbook executable. Default is `ansible-playbook` in PATH | No |
| inventory | string | Default inventory file path (relative to application directory) | No |
| vault | string | Default vault password file path (relative to application directory) | No |
| deployTargets | [][DeployTargetConfig](#DeployTargetConfig) | The config for the destinations to deploy applications | Yes |

#### DeployTargetConfig

| Field | Type | Description | Required |
|-|-|-|-|
| name | string | The name of the deploy target. | Yes |
| labels | map[string]string | The labels of the deploy target. | No |
| config | [AnsibleDeployTargetConfig](#AnsibleDeployTargetConfig) | The configuration of the deploy target for Ansible plugin. | No |

##### AnsibleDeployTargetConfig

| Field | Type | Description | Required |
|-|-|-|-|
| ansiblePath | string | Path to ansible-playbook executable. Overrides plugin-level setting. | No |
| inventory | string | Default inventory file path. Overrides plugin-level setting. | No |
| vault | string | Default vault password file path. Overrides plugin-level setting. | No |

> **Configuration Hierarchy**: Settings are applied in the following order (highest to lowest priority):
> 1. Application-level playbook configuration
> 2. Deploy target configuration  
> 3. Plugin-level configuration
> 4. Default values

### Application Config

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    ansible: # same name as the one defined in `spec.plugins[].name`
      name: my-ansible-app
      playbook:
        path: playbooks/deploy.yml
        inventory: inventory/hosts
        extraVars:
          env: production
          version: "1.0.0"
        tags:
          - deploy
          - configure
        verbosity: 1
        diffMode: true
      pipeline:
        stages:
          - name: ANSIBLE_SYNC
            with:
              timeout: 600
```

| Field | Type | Description | Required |
|-|-|-|-|
| name | string | The name of the application. | No |
| playbook | [AnsiblePlaybookManifest](#AnsiblePlaybookManifest) | The Ansible playbook configuration. | Yes |
| pipeline | [AnsiblePipelineSpec](#AnsiblePipelineSpec) | Pipeline configuration for structured deployments. | No |

#### AnsiblePlaybookManifest

| Field | Type | Description | Required |
|-|-|-|-|
| path | string | Path to the Ansible playbook file (relative to application directory). | Yes |
| inventory | string | Inventory file path. Highest priority - overrides deploy target and plugin settings. | No |
| extraVars | map[string]string | Extra variables to pass to the playbook. | No |
| tags | []string | Tags to run. Only tasks with these tags will be executed. | No |
| skipTags | []string | Tags to skip. Tasks with these tags will be skipped. | No |
| limit | string | Limit execution to specific hosts or groups. | No |
| verbosity | int | Verbosity level (0-4). Higher values provide more output. | No |
| checkMode | bool | Run in check mode (dry-run). No changes will be made. | No |
| diffMode | bool | Show diffs of changes. Useful for reviewing what will be changed. | No |
| vault | string | Vault password file path. Highest priority - overrides deploy target and plugin settings. | No |
| privateKey | string | SSH private key file path for authentication. | No |
| remoteUser | string | Remote user for SSH connections. | No |
| becomeUser | string | User to become (sudo) during playbook execution. | No |
| timeout | int | Timeout in seconds for playbook execution. | No |

#### AnsiblePipelineSpec

| Field | Type | Description | Required |
|-|-|-|-|
| stages | [][AnsibleStageSpec](#AnsibleStageSpec) | List of stages in the pipeline. | No |

##### AnsibleStageSpec

| Field | Type | Description | Required |
|-|-|-|-|
| name | string | Stage name. Must be `ANSIBLE_SYNC`. | Yes |
| with | map[string]interface{} | Stage-specific configuration. | No |

### Stage Config

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    ansible:
      kind: Playbook
      playbook:
        path: playbooks/deploy.yml
      pipeline:
        stages:
          - name: ANSIBLE_SYNC
            with:
              timeout: 600
```

#### `ANSIBLE_SYNC`

| Field | Type | Description | Required |
|-|-|-|-|
| timeout | int | Timeout in seconds for this stage. Overrides playbook-level timeout. | No |



## Usage Examples

### Basic Playbook Execution

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    ansible:
      playbook:
        path: playbooks/deploy.yml
        inventory: inventory/hosts
        verbosity: 1
```

### Advanced Configuration with Variables and Tags

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    ansible:
      playbook:
        path: playbooks/deploy.yml
        inventory: inventory/production
        extraVars:
          app_version: "{{ .Input.Tag }}"
          environment: production
          database_host: "prod-db.example.com"
        tags:
          - deploy
          - configure
        skipTags:
          - debug
        verbosity: 2
        diffMode: true
        checkMode: false
```

### Pipeline Configuration

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    ansible:
      playbook:
        path: playbooks/deploy.yml
        inventory: inventory/hosts
        extraVars:
          deployment_id: "{{ .Input.Trigger.Commit.Hash }}"
          environment: "production"
      pipeline:
        stages:
          - name: ANSIBLE_SYNC
            with:
              timeout: 900
```

### Vault and SSH Key Configuration

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    ansible:
      playbook:
        path: playbooks/deploy.yml
        inventory: inventory/secure-hosts
        vault: vault/production.key
        privateKey: keys/deploy.pem
        remoteUser: deploy
        becomeUser: root
        extraVars:
          secure_mode: "true"
```

## Installation

### Prerequisites

- Ansible must be installed on the system where the plugin runs
- SSH access to target hosts configured
- Proper inventory and playbook files in your application repository

### Building the Plugin

```bash
# Clone the repository
git clone https://github.com/pipe-cd/pipecd-community-plugins.git
cd pipecd-community-plugins/plugins/ansible

# Build for current platform
make build

# Build for Linux (production)
make build-linux

# Run tests
make test

# Clean build artifacts
make clean
```

### Installation Steps

1. **Build the plugin binary:**
   ```bash
   make build-linux
   ```

2. **Deploy the plugin binary to your PipeCD environment:**
   ```bash
   cp ansible-plugin-linux /path/to/piped/plugins/
   ```

3. **Configure the plugin in your piped configuration:**
   ```yaml
   apiVersion: pipecd.dev/v1beta1
   kind: Piped
   spec:
     plugins:
     - name: ansible
       url: file:///path/to/piped/plugins/ansible-plugin
       config:
         ansiblePath: /usr/bin/ansible-playbook
       deployTargets:
         - name: production
           config:
             inventory: inventory/production
   ```

4. **Ensure Ansible is installed on the target system:**
   ```bash
   # Ubuntu/Debian
   sudo apt update && sudo apt install ansible

   # CentOS/RHEL
   sudo yum install ansible

   # macOS
   brew install ansible
   ```

## Contributing

This plugin is part of the [PipeCD Community Plugins](https://github.com/pipe-cd/pipecd-community-plugins) repository. Contributions are welcome!

### Development

```bash
# Install dependencies
go mod download

# Run tests
make test

# Build plugin
make build

# Run integration tests (requires Ansible)
make integration-test
```

### Submitting Changes

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

For more information, see the [Contributing Guide](https://github.com/pipe-cd/pipecd-community-plugins/blob/main/CONTRIBUTING.md).
