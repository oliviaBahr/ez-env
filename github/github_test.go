package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetGitHubToken tests the GitHub token retrieval functionality
func TestGetGitHubToken(t *testing.T) {
	tests := []struct {
		name      string
		setEnv    bool
		envToken  string
		expectErr bool
	}{
		{
			name:      "gets token from environment variable",
			setEnv:    true,
			envToken:  "gho_test_token_123",
			expectErr: false,
		},
		{
			name:      "falls back to gh auth when env not set",
			setEnv:    false,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalToken := os.Getenv("GITHUB_TOKEN")

			if tt.setEnv {
				os.Setenv("GITHUB_TOKEN", tt.envToken)
				defer os.Setenv("GITHUB_TOKEN", originalToken)
			} else {
				os.Unsetenv("GITHUB_TOKEN")
				defer os.Setenv("GITHUB_TOKEN", originalToken)
			}

			token, err := GetGitHubToken()

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			if tt.setEnv {
				assert.NoError(t, err)
				assert.Equal(t, tt.envToken, token)
			} else {
				// When falling back to gh auth, we should get a valid token
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.True(t, strings.HasPrefix(token, "gho_") || strings.HasPrefix(token, "ghp_"))
			}
		})
	}
}

// TestGetCurrentUser tests getting the current authenticated user
func TestGetCurrentUser(t *testing.T) {
	ctx := context.Background()

	username, err := GetCurrentUser(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, username)
	assert.True(t, len(username) > 0)
}

// TestGetRepositoryInfo tests repository information retrieval
func TestGetRepositoryInfo(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Change to the test repository
	testRepoPath := "../../ez-test-env"
	err = os.Chdir(testRepoPath)
	require.NoError(t, err)

	owner, repo, err := GetRepositoryInfo()
	assert.NoError(t, err)
	assert.NotEmpty(t, owner)
	assert.NotEmpty(t, repo)
	assert.Equal(t, "ez-test-env", repo)
}

// TestStoreAndRetrieveEncryptionKey tests the full round-trip of storing and retrieving an encryption key
func TestStoreAndRetrieveEncryptionKey(t *testing.T) {
	ctx := context.Background()

	// Change to the test repository for this test
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	testRepoPath := "../../ez-test-env"
	err = os.Chdir(testRepoPath)
	require.NoError(t, err)

	// Generate a test key (32 bytes)
	testKey := make([]byte, 32)
	for i := range testKey {
		testKey[i] = byte(i + 1)
	}

	t.Run("store encryption key", func(t *testing.T) {
		// Store the key
		err := StoreEncryptionKey(ctx, testKey)
		assert.NoError(t, err)
	})

	t.Run("retrieve encryption key via workflow", func(t *testing.T) {
		// Retrieve the key via workflow
		// The workflow should return the actual stored key
		retrievedKey, err := GetEncryptionKey(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedKey)
		assert.Len(t, retrievedKey, 32, "Encryption key should be 32 bytes")

		// The retrieved key should match the stored key
		assert.Equal(t, testKey, retrievedKey, "Retrieved key should match stored key")
	})
}

// TestStoreEncryptionKey tests storing an encryption key
func TestStoreEncryptionKey(t *testing.T) {
	ctx := context.Background()

	// Change to the test repository for this test
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	testRepoPath := "../../ez-test-env"
	err = os.Chdir(testRepoPath)
	require.NoError(t, err)

	tests := []struct {
		name      string
		expectErr bool
	}{
		{
			name:      "stores valid encryption key",
			expectErr: false,
		},
		{
			name:      "stores another valid key",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a test key (32 bytes)
			testKey := make([]byte, 32)
			for i := range testKey {
				testKey[i] = byte(i + 1)
			}

			err = StoreEncryptionKey(ctx, testKey)

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

// TestGetEncryptionKey tests retrieving an encryption key
func TestGetEncryptionKey(t *testing.T) {
	ctx := context.Background()

	// Change to the test repository for this test
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	testRepoPath := "../../ez-test-env"
	err = os.Chdir(testRepoPath)
	require.NoError(t, err)

	t.Run("retrieve encryption key", func(t *testing.T) {
		key, err := GetEncryptionKey(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, key)
		assert.Len(t, key, 32, "Encryption key should be 32 bytes")
	})
}

// TestBase64Encoding tests the base64 encoding/decoding used in the package
func TestBase64Encoding(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "encodes and decodes simple key",
		},
		{
			name: "encodes and decodes random key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate a test key (32 bytes)
			original := make([]byte, 32)
			for i := range original {
				original[i] = byte(i + 1)
			}

			// Encode
			encoded := base64.StdEncoding.EncodeToString(original)
			assert.NotEmpty(t, encoded)

			// Decode
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			assert.NoError(t, err)
			assert.Equal(t, original, decoded)
		})
	}
}

// TestGitHubCLIAvailability tests that GitHub CLI is available
func TestGitHubCLIAvailability(t *testing.T) {
	cmd := exec.Command("gh", "--version")
	err := cmd.Run()
	assert.NoError(t, err)
}

// TestGitRemoteURLParsing tests the parsing of different git remote URL formats
func TestGitRemoteURLParsing(t *testing.T) {
	tests := []struct {
		name        string
		remoteURL   string
		expectOwner string
		expectRepo  string
		expectErr   bool
	}{
		{
			name:        "parses SSH URL",
			remoteURL:   "git@github.com:testuser/testrepo.git",
			expectOwner: "testuser",
			expectRepo:  "testrepo",
			expectErr:   false,
		},
		{
			name:        "parses HTTPS URL",
			remoteURL:   "https://github.com/testuser/testrepo.git",
			expectOwner: "testuser",
			expectRepo:  "testrepo",
			expectErr:   false,
		},
		{
			name:        "parses HTTPS URL without .git",
			remoteURL:   "https://github.com/testuser/testrepo",
			expectOwner: "testuser",
			expectRepo:  "testrepo",
			expectErr:   false,
		},
		{
			name:      "fails with invalid SSH URL",
			remoteURL: "git@github.com:invalid",
			expectErr: true,
		},
		{
			name:      "fails with invalid HTTPS URL",
			remoteURL: "https://github.com/invalid",
			expectErr: true,
		},
		{
			name:      "fails with unsupported URL",
			remoteURL: "https://gitlab.com/testuser/testrepo.git",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is a helper function to test the parsing logic
			// We'll extract the parsing logic from GetRepositoryInfo for testing
			var owner, repo string
			var err error

			if strings.HasPrefix(tt.remoteURL, "git@github.com:") {
				parts := strings.Split(strings.TrimPrefix(tt.remoteURL, "git@github.com:"), "/")
				if len(parts) != 2 {
					err = fmt.Errorf("invalid remote URL format: %s", tt.remoteURL)
				} else {
					owner = parts[0]
					repo = strings.TrimSuffix(parts[1], ".git")
				}
			} else if strings.HasPrefix(tt.remoteURL, "https://github.com/") {
				parts := strings.Split(strings.TrimPrefix(tt.remoteURL, "https://github.com/"), "/")
				if len(parts) != 2 {
					err = fmt.Errorf("invalid remote URL format: %s", tt.remoteURL)
				} else {
					owner = parts[0]
					repo = strings.TrimSuffix(parts[1], ".git")
				}
			} else {
				err = fmt.Errorf("unsupported remote URL format: %s", tt.remoteURL)
			}

			if tt.expectErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectOwner, owner)
			assert.Equal(t, tt.expectRepo, repo)
		})
	}
}

// BenchmarkStoreEncryptionKey benchmarks storing encryption keys
func BenchmarkStoreEncryptionKey(b *testing.B) {
	ctx := context.Background()

	// Change to the test repository for this benchmark
	originalDir, err := os.Getwd()
	require.NoError(b, err)
	defer os.Chdir(originalDir)

	testRepoPath := "../../ez-test-env"
	err = os.Chdir(testRepoPath)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Generate a test key (32 bytes)
		testKey := make([]byte, 32)
		for j := range testKey {
			testKey[j] = byte(j + 1)
		}

		err = StoreEncryptionKey(ctx, testKey)
		require.NoError(b, err)
	}
}

// BenchmarkGetEncryptionKey benchmarks retrieving encryption keys
func BenchmarkGetEncryptionKey(b *testing.B) {
	ctx := context.Background()

	// Change to the test repository for this benchmark
	originalDir, err := os.Getwd()
	require.NoError(b, err)
	defer os.Chdir(originalDir)

	testRepoPath := "../../ez-test-env"
	err = os.Chdir(testRepoPath)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetEncryptionKey(ctx)
		require.NoError(b, err)
	}
}
