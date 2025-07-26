package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateEncryptionKey(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"generates valid key"},
		{"generates unique keys"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := GenerateEncryptionKey()
			require.NoError(t, err)
			assert.Len(t, key, keySize)
			assert.NotEqual(t, make([]byte, keySize), key) // Should not be all zeros
		})
	}

	// Test that multiple keys are different
	key1, err := GenerateEncryptionKey()
	require.NoError(t, err)
	key2, err := GenerateEncryptionKey()
	require.NoError(t, err)
	assert.NotEqual(t, key1, key2, "Generated keys should be unique")
}

func TestEncryptFile(t *testing.T) {
	// Use a manually generated test key for unit testing
	testKey, err := GenerateEncryptionKey()
	require.NoError(t, err, "failed to generate test key")

	tests := []struct {
		name      string
		plaintext []byte
		key       []byte
		expectErr string
	}{
		{
			name:      "encrypts simple text",
			plaintext: []byte("Hello, World!"),
			key:       testKey,
			expectErr: "",
		},
		{
			name:      "encrypts empty content",
			plaintext: []byte{},
			key:       testKey,
			expectErr: "",
		},
		{
			name:      "encrypts large content",
			plaintext: bytes.Repeat([]byte("A"), 10000),
			key:       testKey,
			expectErr: "",
		},
		{
			name:      "encrypts binary content",
			plaintext: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
			key:       testKey,
			expectErr: "",
		},
		{
			name:      "fails with wrong key size",
			plaintext: []byte("test"),
			key:       []byte("short"),
			expectErr: "invalid key size",
		},
		{
			name:      "fails with nil key",
			plaintext: []byte("test"),
			key:       nil,
			expectErr: "invalid key size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptFile(tt.plaintext, tt.key)

			if tt.expectErr != "" {
				assert.Error(t, err, "expected error but got none")
				assert.Contains(t, err.Error(), tt.expectErr, "error message should contain expected text")
				return
			}

			assert.NoError(t, err, "encryption should succeed")
			assert.NotNil(t, encrypted, "encrypted data should not be nil")
			assert.Greater(t, len(encrypted), len(tt.plaintext), "encrypted data should be larger than plaintext")
			assert.NotEqual(t, tt.plaintext, encrypted, "encrypted data should not match plaintext")
		})
	}
}

func TestDecryptFile(t *testing.T) {
	// Use a manually generated test key for unit testing
	testKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20}

	// Pre-encrypt test data
	simpleText := []byte("Hello, World!")
	simpleTextEncrypted, err := EncryptFile(simpleText, testKey)
	require.NoError(t, err, "encryption should succeed for test setup")

	emptyContent := []byte{}
	emptyContentEncrypted, err := EncryptFile(emptyContent, testKey)
	require.NoError(t, err, "encryption should succeed for test setup")

	largeContent := bytes.Repeat([]byte("A"), 10000)
	largeContentEncrypted, err := EncryptFile(largeContent, testKey)
	require.NoError(t, err, "encryption should succeed for test setup")

	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	binaryContentEncrypted, err := EncryptFile(binaryContent, testKey)
	require.NoError(t, err, "encryption should succeed for test setup")

	tests := []struct {
		name      string
		encrypted []byte
		key       []byte
		expected  []byte
		expectErr string
	}{
		{
			name:      "decrypts simple text",
			encrypted: simpleTextEncrypted,
			key:       testKey,
			expected:  simpleText,
			expectErr: "",
		},
		{
			name:      "decrypts empty content",
			encrypted: emptyContentEncrypted,
			key:       testKey,
			expected:  emptyContent,
			expectErr: "",
		},
		{
			name:      "decrypts large content",
			encrypted: largeContentEncrypted,
			key:       testKey,
			expected:  largeContent,
			expectErr: "",
		},
		{
			name:      "decrypts binary content",
			encrypted: binaryContentEncrypted,
			key:       testKey,
			expected:  binaryContent,
			expectErr: "",
		},
		{
			name: "fails with corrupted data",
			encrypted: func() []byte {
				corrupted := make([]byte, len(simpleTextEncrypted))
				copy(corrupted, simpleTextEncrypted)
				corrupted[len(corrupted)-1] ^= 0xFF
				return corrupted
			}(),
			key:       testKey,
			expectErr: "failed to decrypt",
		},
		{
			name:      "fails with wrong key size",
			encrypted: simpleTextEncrypted,
			key:       []byte("short"),
			expectErr: "invalid key size",
		},
		{
			name:      "fails with nil key",
			encrypted: simpleTextEncrypted,
			key:       nil,
			expectErr: "invalid key size",
		},
		{
			name:      "fails with empty data",
			encrypted: []byte{},
			key:       testKey,
			expectErr: "encrypted data too short",
		},
		{
			name:      "fails with too short data",
			encrypted: []byte{0x00, 0x00, 0x00, 0x01}, // Only version
			key:       testKey,
			expectErr: "encrypted data too short",
		},
		{
			name:      "fails with unsupported version",
			encrypted: append([]byte{0x00, 0x00, 0x00, 0x02}, make([]byte, nonceSize)...), // Version 2
			key:       testKey,
			expectErr: "unsupported version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := DecryptFile(tt.encrypted, tt.key)

			if tt.expectErr != "" {
				assert.Error(t, err, "expected error but got none")
				assert.Contains(t, err.Error(), tt.expectErr, "error message should contain expected text")
				return
			}

			// Test successful decryption
			assert.NoError(t, err, "decryption should succeed")
			if len(tt.expected) == 0 {
				assert.True(t, decrypted == nil || len(decrypted) == 0, "empty plaintext should result in empty or nil output")
			} else {
				assert.Equal(t, tt.expected, decrypted, "decrypted content should match original plaintext")
			}
		})
	}
}

func TestDecryptFileWithWrongKey(t *testing.T) {
	// Use a manually generated test key for unit testing
	testKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20}
	plaintext := []byte("test content")

	// Encrypt with the correct key
	encrypted, err := EncryptFile(plaintext, testKey)
	require.NoError(t, err, "encryption should succeed for test setup")

	// Generate a random wrong key that is different from the test key
	wrongKey := make([]byte, 32)
	for {
		rand.Read(wrongKey)
		if !bytes.Equal(wrongKey, testKey) {
			break
		}
	}

	// Try to decrypt with the wrong key
	_, err = DecryptFile(encrypted, wrongKey)
	assert.Error(t, err, "decryption with wrong key should always return an error")
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	// Use a manually generated test key for unit testing
	testKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20}

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{"simple text", []byte("Hello, World!")},
		{"empty content", []byte{}},
		{"large content", bytes.Repeat([]byte("A"), 10000)},
		{"binary content", []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}},
		{"unicode content", []byte("Hello, 世界! 🌍")},
		{"newlines", []byte("Line 1\nLine 2\r\nLine 3")},
		{"special chars", []byte("!@#$%^&*()_+-=[]{}|;':\",./<>?")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := EncryptFile(tc.plaintext, testKey)
			require.NoError(t, err)

			// Verify it's encrypted
			assert.NotEqual(t, tc.plaintext, encrypted)
			assert.True(t, IsEncryptedFile(encrypted))

			// Decrypt
			decrypted, err := DecryptFile(encrypted, testKey)
			require.NoError(t, err)

			// Verify round trip
			if len(tc.plaintext) == 0 {
				assert.True(t, decrypted == nil || len(decrypted) == 0)
			} else {
				assert.Equal(t, tc.plaintext, decrypted)
			}
		})
	}
}

func TestIsEncryptedFile(t *testing.T) {
	// Use a manually generated test key for unit testing
	testKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20}

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "identifies encrypted file",
			data: func() []byte {
				encrypted, _ := EncryptFile([]byte("test"), testKey)
				return encrypted
			}(),
			want: true,
		},
		{
			name: "identifies non-encrypted file",
			data: []byte("plain text content"),
			want: false,
		},
		{
			name: "identifies empty file",
			data: []byte{},
			want: false,
		},
		{
			name: "identifies short file",
			data: []byte{0x00, 0x00, 0x00},
			want: false,
		},
		{
			name: "identifies wrong version",
			data: append([]byte{0x00, 0x00, 0x00, 0x02}, make([]byte, 100)...), // Version 2
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEncryptedFile(tt.data)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestEncryptionDeterministic(t *testing.T) {
	// Use a manually generated test key for unit testing
	testKey := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F, 0x20}
	plaintext := []byte("test content")

	// Encrypt the same content multiple times
	encrypted1, err := EncryptFile(plaintext, testKey)
	require.NoError(t, err)

	encrypted2, err := EncryptFile(plaintext, testKey)
	require.NoError(t, err)

	// The encrypted content should be different due to random nonce
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to the same plaintext
	decrypted1, err := DecryptFile(encrypted1, testKey)
	require.NoError(t, err)
	decrypted2, err := DecryptFile(encrypted2, testKey)
	require.NoError(t, err)

	assert.Equal(t, plaintext, decrypted1)
	assert.Equal(t, plaintext, decrypted2)
}

func TestLargeFileEncryption(t *testing.T) {
	key, err := GenerateEncryptionKey()
	require.NoError(t, err)

	// Test with a moderately large file (1MB)
	largeContent := make([]byte, 1024*1024)
	_, err = rand.Read(largeContent)
	require.NoError(t, err)

	// Encrypt
	encrypted, err := EncryptFile(largeContent, key)
	require.NoError(t, err)

	// Verify size
	expectedSize := 4 + nonceSize + len(largeContent) + tagSize
	assert.Len(t, encrypted, expectedSize)

	// Decrypt
	decrypted, err := DecryptFile(encrypted, key)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, largeContent, decrypted)
}

func BenchmarkEncryptFile(b *testing.B) {
	key, err := GenerateEncryptionKey()
	require.NoError(b, err)
	plaintext := bytes.Repeat([]byte("A"), 1024) // 1KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncryptFile(plaintext, key)
		require.NoError(b, err)
	}
}

func BenchmarkDecryptFile(b *testing.B) {
	key, err := GenerateEncryptionKey()
	require.NoError(b, err)
	plaintext := bytes.Repeat([]byte("A"), 1024) // 1KB
	encrypted, err := EncryptFile(plaintext, key)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecryptFile(encrypted, key)
		require.NoError(b, err)
	}
}
