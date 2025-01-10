package generator

import (
	"github.com/mach-composer/mach-composer-cli/internal/config"
	"github.com/mach-composer/mach-composer-cli/internal/config/variable"
	"github.com/mach-composer/mach-composer-cli/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockRenderer struct {
	mock.Mock
}

func (m *mockRenderer) Identifier() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockRenderer) StateKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockRenderer) Backend() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockRenderer) RemoteState() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestRenderRemoteSourcesCompactSources(t *testing.T) {

	rr := new(mockRenderer)
	rr.On("Identifier").Return("site/component").Times(2)
	rr.On("StateKey").Return("component")
	rr.On("RemoteState").Return("remote-state", nil)
	rr.On("Backend").Return("backend", nil)

	rr2 := new(mockRenderer)
	rr2.On("Identifier").Return("site/component-2")
	rr2.On("StateKey").Return("component-2")
	rr2.On("RemoteState").Return("remote-state-2", nil)
	rr.On("Backend").Return("backend-2", nil)

	r := state.NewRepository()
	r.Add(rr)
	r.Add(rr2)
	dup, _ := variable.NewScalarVariable("${component.component.value}")
	dup2, _ := variable.NewScalarVariable("${component.component-2.value}")

	sdup := variable.NewSliceVariable([]variable.Variable{dup2, dup})

	scc := config.SiteComponentConfig{
		Variables: map[string]variable.Variable{
			"value":   dup,
			"value_2": sdup,
		},
	}

	resp, err := renderRemoteSources(r, "site", scc)
	assert.NoError(t, err)
	assert.Equal(t, "remote-state\nremote-state-2", resp)
}
