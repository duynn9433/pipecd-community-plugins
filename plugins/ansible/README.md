# Ansible Plugin

This plugin provides Ansible playbook execution capabilities for PipeCD.

## Features

- Execute Ansible playbooks as part of your deployment pipeline
- Support for inventory files, extra variables, tags, and various Ansible options
- Configurable verbosity and execution modes (check mode, diff mode)
- Vault support for encrypted variables
- SSH key and user configuration

## Configuration

### Plugin Configuration

```yaml
config:
  ansiblePath: /usr/local/bin/ansible-playbook  # Optional: path to ansible-playbook binary
  inventory: inventory/hosts                    # Optional: default inventory file
  vault: vault/password                        # Optional: default vault password file
```

### Stage Configuration

```yaml
stages:
  - name: run-playbook
    with:
      playbook: playbooks/deploy.yml             # Required: path to playbook file
      inventory: inventory/production            # Optional: inventory file (overrides plugin config)
      extraVars:                                # Optional: extra variables
        env: production
        version: "1.0.0"
      tags:                                     # Optional: tags to run
        - deploy
        - configure
      skipTags:                                 # Optional: tags to skip
        - debug
      limit: webservers                         # Optional: limit to specific hosts
      verbosity: 2                              # Optional: verbosity level (0-4)
      checkMode: false                          # Optional: run in check mode
      diffMode: true                            # Optional: show diffs
      vault: vault/prod-password                # Optional: vault password file
      privateKey: keys/deploy.pem               # Optional: SSH private key
      remoteUser: deploy                        # Optional: remote user
      becomeUser: root                          # Optional: become user
      timeout: 600                              # Optional: timeout in seconds
```

## Usage

1. Add the plugin to your PipeCD configuration
2. Configure your pipeline stages to use `ANSIBLE_PLAYBOOK`
3. Ensure ansible-playbook is installed and accessible
4. Place your playbooks, inventory, and other files in your deployment repository

## Example

```yaml
apiVersion: pipecd.dev/v1beta1
kind: KubernetesApp
metadata:
  name: my-app
spec:
  pipeline:
    stages:
      - name: configure-infrastructure
        with:
          playbook: ansible/configure.yml
          inventory: ansible/inventory/production
          extraVars:
            app_version: "{{ .Input.Tag }}"
            environment: production
          tags:
            - configure
            - deploy
          verbosity: 1
```