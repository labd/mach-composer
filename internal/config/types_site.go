package config

import (
	"fmt"
	"log"

	"github.com/creasty/defaults"
	"github.com/elliotchance/pie/v2"
	"gopkg.in/yaml.v3"
)

// Site contains all configuration needed for a site.
type Site struct {
	Name         string
	Identifier   string
	RawEndpoints map[string]any `yaml:"endpoints"`
	Endpoints    []Endpoint     `yaml:"_endpoints"`

	Components []SiteComponent `yaml:"components"`

	AWS   *SiteAWS           `yaml:"aws,omitempty"`
	Azure *SiteAzureSettings `yaml:"azure,omitempty"`

	Commercetools *CommercetoolsSettings `yaml:"commercetools"`
	Amplience     *AmplienceConfig       `yaml:"amplience"`
	Sentry        *SentryConfig          `yaml:"sentry"`
}

func (s *Site) resolveEndpoints() error {
	for k, rv := range s.RawEndpoints {
		switch v := rv.(type) {
		case string:
			ep := Endpoint{
				Key: k,
				URL: v,
			}
			if err := defaults.Set(&ep); err != nil {
				return err
			}
			s.Endpoints = append(s.Endpoints, ep)

		case map[string]any:
			// Do an extra serialize/deserialize step here. Simplest solution
			// for now.

			body, err := yaml.Marshal(v)
			if err != nil {
				return err
			}

			ep := Endpoint{
				Key: k,
			}
			if err := yaml.Unmarshal(body, &ep); err != nil {
				return err
			}

			if err := defaults.Set(&ep); err != nil {
				return err
			}

			s.Endpoints = append(s.Endpoints, ep)
		default:
			return fmt.Errorf("unexecpted type while resolving endpoints")
		}
	}

	// Check if we need to add a default endpoint
	components := s.EndpointComponents()
	keys := make([]string, 0, len(s.Endpoints))
	for _, e := range s.Endpoints {
		keys = append(keys, e.Key)
	}

	// If one of the components has a 'default' endpoint defined,
	// we'll include it to our site endpoints.
	// A 'default' endpoint is one without a custom domain, so no further
	// Route53 or DNS zone settings required.
	componentKeys := []string{}
	for k := range components {
		componentKeys = append(componentKeys, k)
	}
	if stringContains(componentKeys, "default") && stringContains(keys, "default") {
		fmt.Println(
			"WARNING: 'default' endpoint used but not defined in the site endpoints.\n" +
				"MACH will create a default endpoint without any custom domain attached to it.\n" +
				"More info: https://docs.machcomposer.io/reference/syntax/sites.html#endpoints",
		)
		s.Endpoints = append(s.Endpoints, Endpoint{
			URL: "",
			Key: "default",
		})
	}
	return nil
}

func (s *Site) EndpointComponents() map[string][]SiteComponent {
	// Check if we need to add a default endpoint
	endpoints := make(map[string][]SiteComponent)
	for _, c := range s.Components {
		for _, value := range c.Definition.Endpoints {
			endpoints[value] = append(endpoints[value], c)
		}
	}
	return endpoints
}

// UsedEndpoints returns only the endpoints that are actually used by the components.
func (s *Site) UsedEndpoints() []Endpoint {
	result := []Endpoint{}
	for _, ep := range s.Endpoints {
		if len(ep.Components) > 0 {
			result = append(result, ep)
		}
	}
	return result
}

// DNSZones returns the DNS zones of used endpoints.
func (s *Site) DNSZones() []string {
	result := []string{}
	endpoints := s.UsedEndpoints()
	for i := range endpoints {
		result = append(result, endpoints[i].Zone)
	}
	return pie.Unique(result)
}

// HasCDNEndpoint checks if there is an endpoint with a CDN enabled.
func (s *Site) HasCDNEndpoint() bool {
	endpoints := s.UsedEndpoints()
	for _, ep := range endpoints {
		if ep.AWS != nil && ep.AWS.EnableCDN {
			return true
		}
	}
	return false
}

type SiteComponent struct {
	Name      string
	Variables map[string]any
	Secrets   map[string]any

	StoreVariables map[string]any `yaml:"store_variables"`
	StoreSecrets   map[string]any `yaml:"store_secrets"`

	Definition *Component
	Sentry     *SentryConfig `yaml:"sentry"`
}

func (sc SiteComponent) HasCloudIntegration() bool {
	if sc.Definition == nil {
		log.Fatalf("Component %s was not resolved properly (missing definition)", sc.Name)
	}
	for _, i := range sc.Definition.Integrations {
		if i == "aws" || i == "azure" {
			return true
		}
	}
	return false
}

type Endpoint struct {
	URL   string         `yaml:"url"`
	Key   string         `yaml:"key"`
	Zone  string         `yaml:"zone"`
	AWS   *AWSEndpoint   `yaml:"aws"`
	Azure *AzureEndpoint `yaml:"azure"`

	Components []SiteComponent
}

func (e *Endpoint) SetDefaults() {
	e.URL = StripProtocol(e.URL)

	if e.Zone == "" && e.URL != "" {
		e.Zone = ZoneFromURL(e.URL)
	}
}

func (e *Endpoint) IsRootDomain() bool {
	return e.URL == e.Zone
}

func (e Endpoint) Subdomain() string {
	if e.URL == "" {
		return ""
	}

	return SubdomainFromURL(e.URL)
}
