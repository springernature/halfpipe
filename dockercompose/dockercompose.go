package dockercompose

import (
	"regexp"
	"strings"

	"sort"

	"github.com/ghodss/yaml"
	"github.com/spf13/afero"
)

type DockerCompose struct {
	Services []Service
}

func (dc DockerCompose) HasService(name string) bool {
	for _, s := range dc.Services {
		if s.Name == name {
			return true
		}
	}
	return false
}

type Service struct {
	Name  string
	Image string
}

func (s Service) HasImage() bool {
	return s.Image != ""
}

func (s Service) ResourceName() string {
	return "dockercompose-" + regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(strings.ToLower(s.Name), "_")
}

type Reader func() (DockerCompose, error)

var (
	filePath = "docker-compose.yml"
)

func NewReader(fs afero.Afero) Reader {
	return func() (DockerCompose, error) {
		dc := DockerCompose{}
		content, err := fs.ReadFile(filePath)
		if err != nil {
			return dc, err
		}

		var parsed struct {
			Services map[string]map[string]interface{}
		}

		err = yaml.Unmarshal(content, &parsed)
		if err != nil {
			return dc, err
		}

		// try parsing without services key
		if len(parsed.Services) == 0 {
			err = yaml.Unmarshal(content, &parsed.Services)
			if err != nil {
				return dc, err
			}
		}

		for serviceName, serviceMap := range parsed.Services {
			service := Service{Name: serviceName}
			if image, ok := serviceMap["image"].(string); ok {
				service.Image = image
			}
			dc.Services = append(dc.Services, service)
		}

		sort.Slice(dc.Services, func(i, j int) bool {
			return dc.Services[i].Name < dc.Services[j].Name
		})

		return dc, nil
	}
}
