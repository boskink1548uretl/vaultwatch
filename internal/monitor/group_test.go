package monitor

import (
	"bytes"
	"context"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yourusername/vaultwatch/internal/vault"
)

type mockGroupClient struct {
	groups []vault.Group
	err    error
}

func (m *mockGroupClient) ListGroups(_ context.Context) ([]vault.Group, error) {
	return m.groups, m.err
}

func (m *mockGroupClient) GetGroup(_ context.Context, _ string) (*vault.Group, error) {
	return nil, nil
}

func newGroupLogger() (*log.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return log.New(buf, "", 0), buf
}

func TestGroupWatcher_NoRiskyPolicies(t *testing.T) {
	client := &mockGroupClient{
		groups: []vault.Group{
			{ID: "g1", Name: "dev-team", Policies: []string{"default", "dev"}},
		},
	}
	logger, _ := newGroupLogger()
	w := NewGroupWatcher(client, nil, logger)

	risks, err := w.Audit(context.Background())
	require.NoError(t, err)
	assert.Empty(t, risks)
}

func TestGroupWatcher_DetectsRiskyPolicy(t *testing.T) {
	client := &mockGroupClient{
		groups: []vault.Group{
			{ID: "g2", Name: "ops-team", Policies: []string{"default", "root"}},
		},
	}
	logger, buf := newGroupLogger()
	w := NewGroupWatcher(client, []string{"root", "admin"}, logger)

	risks, err := w.Audit(context.Background())
	require.NoError(t, err)
	require.Len(t, risks, 1)
	assert.Equal(t, "ops-team", risks[0].GroupName)
	assert.Equal(t, "root", risks[0].RiskyPolicy)
	assert.Contains(t, buf.String(), "root")
}

func TestGroupWatcher_MultipleGroups(t *testing.T) {
	client := &mockGroupClient{
		groups: []vault.Group{
			{ID: "g1", Name: "safe-team", Policies: []string{"read-only"}},
			{ID: "g2", Name: "admin-team", Policies: []string{"admin"}},
			{ID: "g3", Name: "super-team", Policies: []string{"superuser"}},
		},
	}
	logger, _ := newGroupLogger()
	w := NewGroupWatcher(client, nil, logger)

	risks, err := w.Audit(context.Background())
	require.NoError(t, err)
	assert.Len(t, risks, 2)
}

func TestGroupWatcher_ClientError(t *testing.T) {
	client := &mockGroupClient{err: errors.New("vault unavailable")}
	logger, _ := newGroupLogger()
	w := NewGroupWatcher(client, nil, logger)

	_, err := w.Audit(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vault unavailable")
}

func TestGroupWatcher_DefaultThreshold(t *testing.T) {
	w := NewGroupWatcher(&mockGroupClient{}, nil, nil)
	assert.Equal(t, []string{"root", "admin", "superuser"}, w.riskyPolicies)
}
