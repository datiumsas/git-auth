package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/crypto/ssh"
)

type SSHManager struct {
	host    string
	keyName string
	path    string
	port    int
}

// New creates a new instance of SSHManager with optional configurations.
func New(host string, ops ...Options) *SSHManager {
	sshConfig := &SSHManager{
		host:    host,
		port:    22,
		path:    "~/.ssh",
		keyName: "id_rsa",
	}

	for _, op := range ops {
		op(sshConfig)
	}
	return sshConfig
}

// AddSSHConfig generates and appends or updates the SSH config.
func (cfg *SSHManager) AddSSHConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	sshConfigPath := filepath.Join(home, ".ssh", "config")
	if err := os.MkdirAll(filepath.Dir(sshConfigPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// SSH Config Template
	templateFile := `Host {{.Host}}
    HostName {{.HostName}}
    Port {{.Port}}
    IdentityFile {{.IdentityFile}}
`
	// Set up SSH config details
	sshConfig := struct {
		Host         string
		HostName     string
		Port         int
		IdentityFile string
	}{
		Host:         cfg.host,
		HostName:     cfg.host,
		Port:         cfg.port,
		IdentityFile: filepath.Join(cfg.path, cfg.keyName),
	}

	// Parse the embedded template
	tmpl, err := template.New("sshConfig").Parse(templateFile)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	// Apply the template to generate the config
	var sb strings.Builder
	if err := tmpl.Execute(&sb, sshConfig); err != nil {
		return fmt.Errorf("error applying template: %v", err)
	}

	// Create the new configuration to insert into the file
	newConfig := sb.String()

	// Read the existing SSH config file (if exists)
	sshConfigFile, err := os.ReadFile(sshConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading SSH config file: %v", err)
	}

	// Convert file content to string for easier processing
	sshConfigContent := string(sshConfigFile)

	// Define a unique header using the host as an identifier
	headerTag := fmt.Sprintf("# BEGIN GENERATED CONFIG FOR %s", cfg.host)
	footerTag := fmt.Sprintf("# END GENERATED CONFIG FOR %s", cfg.host)

	// Check if the file already contains a configuration for this host
	if strings.Contains(sshConfigContent, headerTag) {
		// Find the start and end of the existing section for the host
		start := strings.Index(sshConfigContent, headerTag)
		end := strings.Index(sshConfigContent, footerTag)

		if start >= 0 && end > start {
			// Replace the existing section with the new configuration
			sshConfigContent = sshConfigContent[:start+len(headerTag)] + "\n" + newConfig + "\n" + sshConfigContent[end:]
		}
	} else {
		// If the section for the host doesn't exist, add it with the new configuration
		sshConfigContent += "\n" + headerTag + "\n" + newConfig + "\n" + footerTag + "\n"
	}

	// Write the updated SSH config back to the file
	err = os.WriteFile(sshConfigPath, []byte(sshConfigContent), 0644)
	if err != nil {
		return fmt.Errorf("error writing to SSH config file: %v", err)
	}

	return nil
}

// GenerateSSHKeyPair generates an SSH key pair (private and public).
func (cfg *SSHManager) GenerateSSHKeyPair() (privateKeyPath, publicKeyPath string, err error) {
	// Ensure the SSH path exists, if not create it
	if _, err := os.Stat(cfg.path); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.path, os.ModePerm); err != nil {
			return "", "", fmt.Errorf("failed to create SSH directory: %w", err)
		}
	}

	// Define the paths for the private and public keys
	privateKeyPath = filepath.Join(cfg.path, cfg.keyName)
	publicKeyPath = privateKeyPath + ".pub"

	// Generate an RSA private key (4096-bit key for better security)
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	// Write the private key to a file in PEM format
	privateKeyFile, err := os.Create(privateKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create private key file: %w", err)
	}
	defer privateKeyFile.Close()

	// Encode the private key in PEM format
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	err = pem.Encode(privateKeyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKeyBytes})
	if err != nil {
		return "", "", fmt.Errorf("failed to encode private key: %w", err)
	}

	// Set permissions to 600 for the private key
	if err := os.Chmod(privateKeyPath, 0600); err != nil {
		return "", "", fmt.Errorf("failed to set permissions on private key: %w", err)
	}

	// Generate the corresponding public key
	publicKey := &privateKey.PublicKey

	// Convert public key to the SSH format
	sshPubKey, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to convert public key to SSH format: %w", err)
	}

	// Write the public key to the file in the proper SSH format
	publicKeyFile, err := os.Create(publicKeyPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create public key file: %w", err)
	}
	defer publicKeyFile.Close()

	_, err = publicKeyFile.Write(ssh.MarshalAuthorizedKey(sshPubKey))
	if err != nil {
		return "", "", fmt.Errorf("failed to write public key to file: %w", err)
	}

	return privateKeyPath, publicKeyPath, nil
}
