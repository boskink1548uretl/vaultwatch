package monitor

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/vaultwatch/internal/vault"
)

type mockProfileClient struct {
	profile *vault.VaultProfile
	err     error
}

func (m *mockProfileClient) GetProfile(_ context.Context) (*vault.VaultProfile, error) {
	return m.profile, m.err
}

func newProfileWatcher(client ProfileClient) (*ProfileWatcher, *bytes.Buffer) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)
	w := NewProfileWatcher(client, WithProfileLogger(logger))
	return w, &logBuf
}

func TestProfileWatcher_HealthyVault(t *testing.T) {
	client := &mockProfileClient{
		profile: &vault.VaultProfile{
			Version: "1.15.0", ClusterName: "prod", ClusterID: "xyz",
			HA: true, Sealed: false, Initialized: true,
		},
	}
	watcher, logBuf := newProfileWatcher(client)
	var out bytes.Buffer
	err := watcher.Check(context.Background(), &out)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "1.15.0")
	assert.Contains(t, out.String(), "prod")
	assert.Contains(t, logBuf.String(), "Vault healthy")
}

func TestProfileWatcher_SealedVault(t *testing.T) {
	client := &mockProfileClient{
		profile: &vault.VaultProfile{
			Version: "1.14.0", Sealed: true, Initialized: true,
		},
	}
	watcher, logBuf := newProfileWatcher(client)
	var out bytes.Buffer
	err := watcher.Check(context.Background(), &out)
	require.NoError(t, err)
	assert.True(t, strings.Contains(logBuf.String(), "WARNING"))
	assert.Contains(t, out.String(), "true")
}

func TestProfileWatcher_ClientError(t *testing.T) {
	client := &mockProfileClient{err: errors.New("connection refused")}
	watcher, _ := newProfileWatcher(client)
	var out bytes.Buffer
	err := watcher.Check(context.Background(), &out)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "profile watcher")
}

func TestProfileWatcher_DefaultThreshold(t *testing.T) {
	client := &mockProfileClient{
		profile: &vault.VaultProfile{Version: "1.15.0", HA: false},
	}
	w := NewProfileWatcher(client)
	assert.NotNil(t, w)
	assert.NotNil(t, w.logger)
}
