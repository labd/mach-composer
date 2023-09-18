package state

import (
	"github.com/mach-composer/mach-composer-cli/internal/utils"
)

type TerraformCloudState struct {
	Hostname     string `mapstructure:"hostname"`
	Organization string `mapstructure:"organization"`
	Token        string `mapstructure:"token"`
	Workspaces   struct {
		Name   string `mapstructure:"name"`
		Prefix string `mapstructure:"prefix"`
	} `mapstructure:"workspaces"`
}

type TerraformCloudRenderer struct {
	key   string
	state *TerraformCloudState
}

func (tcr *TerraformCloudRenderer) Key() string {
	return tcr.key
}

func (tcr *TerraformCloudRenderer) Backend() (string, error) {
	templateContext := struct {
		State *TerraformCloudState
		Key   string
	}{
		State: tcr.state,
		Key:   tcr.key,
	}

	tpl := `
	backend "remote" {
	  organization = "{{ .State.Organization }}"
	  {{ if .State.Hostname }}
	  hostname = "{{ .State.Hostname }}"
	  {{ end }}
	  {{ if .State.Workspaces }}
	  workspaces {
		{{ if .State.Workspaces.Name }}
		name = "{{ .State.Workspaces.Name }}"
	    {{ end }}
		{{ if .State.Workspaces.Prefix }}
	    prefix = "{{ .State.Workspaces.Prefix }}"
	    {{ end }}
	  }
	  {{ end }}
	}
	`
	return utils.RenderGoTemplate(tpl, templateContext)
}

func (tcr *TerraformCloudRenderer) RemoteState() (string, error) {
	templateContext := struct {
		State *TerraformCloudState
		Key   string
	}{
		State: tcr.state,
		Key:   tcr.key,
	}

	template := `
	data "terraform_remote_state" "{{ .Key }}" {
	  backend = "remote"
	
	  config = {
	    organization = "{{ .State.Organization }}"
	    {{ if .State.Hostname }}
	    hostname = "{{ .State.Hostname }}"
	    {{ end }}
	    {{ if .State.Workspaces }}
	    workspaces {
		  {{ if .State.Workspaces.Name }}
		  name = "{{ .State.Workspaces.Name }}"
	      {{ end }}
		  {{ if .State.Workspaces.Prefix }}
	      prefix = "{{ .State.Workspaces.Prefix }}"
	      {{ end }}
	    }
	    {{ end }}
	  }
	}
	`
	return utils.RenderGoTemplate(template, templateContext)
}
