package store

import (
	"errors"
	"slices"
	"strings"

	"github.com/ml-ai-ops/platform/pkg/api"
)

var validServices = []string{
	"overview", "projects", "pipelines", "models", "agents", "features",
	"storage", "realtime", "catalog", "platform", "workbench", "ide",
}

func validateAccess(subject string, req api.UpsertUserAccessRequest) (api.UserAccess, error) {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return api.UserAccess{}, errors.New("subject is required")
	}
	if req.Role == "" {
		req.Role = "user"
	}
	if req.Role != "admin" && req.Role != "user" {
		return api.UserAccess{}, errors.New("role must be admin or user")
	}
	req.Services = unique(req.Services)
	req.ProjectIDs = unique(req.ProjectIDs)
	req.Storage.Buckets = unique(req.Storage.Buckets)
	for _, service := range req.Services {
		if !slices.Contains(validServices, service) {
			return api.UserAccess{}, errors.New("unknown service: " + service)
		}
	}
	if req.Storage.SizeGB < 0 || req.Compute.VCPUs < 0 || req.Compute.MemoryGB < 0 ||
		req.Compute.MaxVMs < 0 || req.Compute.MaxProjects < 0 || req.Compute.MaxRuns < 0 {
		return api.UserAccess{}, errors.New("resource limits cannot be negative")
	}
	return api.UserAccess{
		Subject: subject, Email: strings.TrimSpace(req.Email), Role: req.Role,
		Services: req.Services, ProjectIDs: req.ProjectIDs, Storage: req.Storage,
		Compute: req.Compute, Disabled: req.Disabled,
	}, nil
}

func unique(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	return result
}
