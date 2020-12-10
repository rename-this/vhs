package smoke

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

// Config is configuration for a smoke test.
type Config struct {
	ReadinessPath string
	ReadinessPort string
	RunOptions    *dockertest.RunOptions
}

//

// SetupDockertestPool creates a new dockertest pool.
func SetupDockertestPool(cfg Config) (*dockertest.Resource, func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("failed to create docker pool: %v", err)
	}

	resource, err := pool.RunWithOptions(cfg.RunOptions, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		log.Fatalf("failed to run with options: %v", err)
	}

	if cfg.ReadinessPath != "" {
		u := url.URL{
			Scheme: "http",
			Host:   "localhost:" + resource.GetPort(cfg.ReadinessPort+"/tcp"),
			Path:   cfg.ReadinessPath,
		}

		err := pool.Retry(func() error {
			res, err := http.Get(u.String())
			if err != nil {
				return err
			}
			if res.StatusCode < 200 || res.StatusCode > 299 {
				return errors.New("bad status")
			}
			return nil
		})
		if err != nil {
			log.Fatalf("failed readiness check: %v", err)
		}
	}

	return resource, func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("failed to purge: %v", err)
		}
	}
}
