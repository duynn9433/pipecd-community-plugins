# ansible plugin

Deploy applications using Ansible playbooks with PipeCD.

## Quick Start

Create an application configuration:

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

## Configuration

### Piped Configuration

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Piped
spec:
  plugins:
  - name: ansible
    url: file:///path/to/ansible-plugin
    deployTargets:
      - name: production
        config:
          ansiblePath: /usr/bin/ansible-playbook
          inventory: inventory/production
```

### Application Configuration

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    ansible:
      playbook:
        path: playbooks/deploy.yml           # Required: Path to playbook
        inventory: inventory/hosts           # Inventory file
        extraVars:                          # Variables to pass
          env: production
          version: "1.0.0"
        tags: [deploy, configure]           # Run only these tags
        verbosity: 1                        # Verbosity level (0-4)
        checkMode: false                    # Dry-run mode
        diffMode: true                      # Show diffs
        timeout: 600                        # Timeout in seconds
```

### Pipeline Configuration

For structured deployments with multiple stages:

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    ansible:
      playbook:
        path: playbooks/deploy.yml
        inventory: inventory/hosts
      pipeline:
        stages:
          - name: ANSIBLE_SYNC
            with:
              timeout: 900
```

## Configuration Reference

### Deploy Target Configuration Keys

Configure in piped config under `deployTargets[].config`:

| Key | Type | Description | Required |
|-----|------|-------------|----------|
| `ansiblePath` | string | Path to ansible-playbook executable | No |
| `inventory` | string | Default inventory file path | No |
| `vault` | string | Default vault password file path | No |

### Application Configuration Keys

Configure in application config under `spec.plugins.ansible.playbook`:

| Key | Type | Description | Required | Default |
|-----|------|-------------|----------|---------|
| `path` | string | Path to the Ansible playbook file | Yes | - |
| `inventory` | string | Inventory file path | No | - |
| `extraVars` | map[string]string | Extra variables to pass to playbook | No | - |
| `tags` | []string | Tags to run (only tasks with these tags) | No | - |
| `skipTags` | []string | Tags to skip | No | - |
| `limit` | string | Limit execution to specific hosts/groups | No | - |
| `verbosity` | int | Verbosity level (0-4) | No | 0 |
| `checkMode` | bool | Run in check mode (dry-run) | No | false |
| `diffMode` | bool | Show diffs of changes | No | false |
| `vault` | string | Vault password file path | No | - |
| `privateKey` | string | SSH private key file path | No | - |
| `remoteUser` | string | Remote user for SSH connections | No | - |
| `becomeUser` | string | User to become (sudo) during execution | No | - |
| `timeout` | int | Timeout in seconds for execution | No | 0 (no timeout) |

### Pipeline Stage Configuration Keys

Configure in pipeline stage under `stages[].with`:

| Key | Type | Description | Required | Default |
|-----|------|-------------|----------|---------|
| `timeout` | int | Timeout in seconds for this stage | No | Uses playbook timeout |

## Examples

### Basic Deployment

```yaml
apiVersion: pipecd.dev/v1beta1
kind: Application
spec:
  plugins:
    ansible:
      playbook:
        path: deploy.yml
        inventory: hosts
```

### Advanced Configuration

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
        tags: [deploy, configure]
        verbosity: 2
        diffMode: true
        vault: vault/production.key
        privateKey: keys/deploy.pem
        remoteUser: deploy
        becomeUser: root
```

## Installation

### Prerequisites

- Ansible installed on the system
- SSH access to target hosts
- Playbook and inventory files in your repository

### Setup

**Configure in piped:**
   ```yaml
   apiVersion: pipecd.dev/v1beta1
   kind: Piped
   spec:
     plugins:
     - name: ansible
       url: file:///path/to/piped/plugins/ansible-plugin
   ```
