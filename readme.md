# GIT-Auth CLI

A command-line tool for managing authentication and SSH keys in GitLab. This tool simplifies authentication, cleans up old SSH keys, and automates the addition of new keys to GitLab.

## Features

- **Authentication:** Authenticate with a GitLab instance using device flow.
- **SSH Key Management:**
  - Clean old SSH keys from GitLab by title prefix.
  - Generate and upload new SSH keys to GitLab.
- **Combined Workflow:** A single command to authenticate, clean old keys, and add a new key.
- **SSH Configuration Management:**
  - Generate and append SSH configuration for a specified host to the user's `~/.ssh/config`.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/atnomoverflow/git-auth-cli.git
   cd git-auth-cli
   ```

2. Build the project:
   ```bash
   go build -o git-auth
   ```

3. Add the binary to your PATH (optional):
   ```bash
   export PATH=$PATH:/path/to/git-auth
   ```

## Configuration

The tool requires a configuration file stored in the `.git-auth` folder under the home directory. The file should be named `config`.

### Example Configuration

```ini
[default]
url = "http://localhost:8000"
ssh-host = "localhost"
ssh-port = "2222"
ssh-prefix = "atnomoverflow"
ssh-path = "~/.git-auth-ssh"
client-id = "84c65c6208a18419df288dd5c29deeb1f270957d38f764414face2afa07b8947"
scope = ["api", "write_repository", "read_user"]
```

- **Fields:**
  - `url`: The URL of your GitLab instance.
  - `ssh-host`: Host for SSH connections.
  - `ssh-port`: Port for SSH connections.
  - `ssh-prefix`: Prefix used for managing SSH keys.
  - `ssh-path`: Path to store SSH keys locally.
  - `client-id`: The GitLab application client ID.
  - `scope`: List of permissions required for the tool.

Ensure this file is present in `~/.git-auth/config` before using the tool.

## Usage

### Commands

#### 1. `auth`
Authenticate with the specified GitLab instance.

- **Usage:**
  ```bash
  git-auth auth
  ```
- **Description:** Authenticates the user by fetching or refreshing the token and validates the login.

---

#### 2. `clean-keys`
Remove SSH keys from GitLab by title prefix.

- **Usage:**
  ```bash
  git-auth clean-keys [--prefix <prefix>]
  ```
- **Options:**
  - `--prefix`: Specify a custom prefix for keys to clean. Defaults to the prefix defined in the configuration.

---

#### 3. `add-key`
Generate and add a new SSH key to GitLab.

- **Usage:**
  ```bash
  git-auth add-key
  ```
- **Description:** Generates a new SSH key pair and uploads the public key to the authenticated GitLab account. The private key is saved locally.

---

#### 4. `magic-auth`
Authenticate, clean old keys, and add a new key to GitLab in a single command.

- **Usage:**
  ```bash
  git-auth magic-auth 
  ```
- **Description:** Combines the functionality of `auth`, `clean-keys`, and `add-key`. Automatically handles login, removes old SSH keys, and uploads a new one.

---

#### 5. `generate-ssh-config`
Generates and appends SSH configuration for a specified host.

- **Usage:**
  ```bash
  git-auth generate-ssh-config
  ```
- **Description:** This command generates an SSH configuration file for the provided host if one does not already exist. It adds the host, hostname, port, and identity file details into the user's SSH config file located at `~/.ssh/config`. If the configuration already exists, it will update the configuration for the host rather than appending.

---

## Examples

1. **Authenticate with GitLab:**
   ```bash
   git-auth auth
   ```

2. **Remove SSH keys with a specific prefix:**
   ```bash
   git-auth clean-keys --prefix my-prefix
   ```

3. **Add a new SSH key to GitLab:**
   ```bash
   git-auth add-key
   ```

4. **Run the full workflow with `magic-auth`:**
   ```bash
   git-auth magic-auth --prefix my-prefix
   ```

5. **Generate and append SSH configuration for a host:**
   ```bash
   git-auth generate-ssh-config
   ```

---

## Contributing

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a clear description of your changes.

---

## Contact

For questions or feedback, reach out to:

**Montasser Abed Majid Zehri**  
📧 montasser.zehri@gmail.com
