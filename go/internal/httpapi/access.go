package httpapi

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/ml-ai-ops/platform/internal/auth"
	"github.com/ml-ai-ops/platform/internal/store"
	"github.com/ml-ai-ops/platform/pkg/api"
)

func principal(r interface{ Context() context.Context }) auth.Principal {
	value, _ := auth.PrincipalFrom(r.Context())
	return value
}

func privileged(value auth.Principal) bool {
	return slices.Contains(value.Roles, auth.RoleAdmin) ||
		slices.Contains(value.Roles, auth.RoleOperator) ||
		slices.Contains(value.Roles, auth.RoleService)
}

func accessFor(repository store.Repository, value auth.Principal) any {
	if privileged(value) {
		return nil
	}
	access, err := repository.AccessFor(value.Subject)
	if err != nil {
		return nil
	}
	return access
}

func allowedProjectIDs(repository store.Repository, value auth.Principal) map[string]bool {
	if privileged(value) || !slices.Contains(value.Roles, auth.RoleUser) {
		return nil
	}
	allowed := make(map[string]bool, len(value.ProjectIDs))
	for _, id := range value.ProjectIDs {
		allowed[id] = true
	}
	for _, project := range repository.Projects() {
		if project.OwnerSubject == value.Subject {
			allowed[project.ID] = true
		}
	}
	return allowed
}

func projectAllowed(repository store.Repository, value auth.Principal, projectID string) bool {
	allowed := allowedProjectIDs(repository, value)
	return allowed == nil || allowed[projectID]
}

func filterProjects(items []api.Project, value auth.Principal) []api.Project {
	if !slices.Contains(value.Roles, auth.RoleUser) {
		return items
	}
	allowed := make(map[string]bool, len(value.ProjectIDs))
	for _, id := range value.ProjectIDs {
		allowed[id] = true
	}
	result := make([]api.Project, 0)
	for _, item := range items {
		if allowed[item.ID] || item.OwnerSubject == value.Subject {
			result = append(result, item)
		}
	}
	return result
}

func filterRuns(items []api.PipelineRun, allowed map[string]bool) []api.PipelineRun {
	if allowed == nil {
		return items
	}
	result := make([]api.PipelineRun, 0)
	for _, item := range items {
		if allowed[item.ProjectID] {
			result = append(result, item)
		}
	}
	return result
}

func filterModels(items []api.Model, allowed map[string]bool) []api.Model {
	if allowed == nil {
		return items
	}
	result := make([]api.Model, 0)
	for _, item := range items {
		if allowed[item.ProjectID] {
			result = append(result, item)
		}
	}
	return result
}

func filterAgents(items []api.Agent, allowed map[string]bool) []api.Agent {
	if allowed == nil {
		return items
	}
	result := make([]api.Agent, 0)
	for _, item := range items {
		if allowed[item.ProjectID] {
			result = append(result, item)
		}
	}
	return result
}

type scopedCatalog struct {
	repository store.Repository
	allowed    map[string]bool
}

func (s scopedCatalog) Models() []api.Model             { return filterModels(s.repository.Models(), s.allowed) }
func (s scopedCatalog) Agents() []api.Agent             { return filterAgents(s.repository.Agents(), s.allowed) }
func (s scopedCatalog) Tools() []api.Tool               { return s.repository.Tools() }
func (s scopedCatalog) FeatureViews() []api.FeatureView { return s.repository.FeatureViews() }

func modelAllowed(repository store.Repository, value auth.Principal, modelID string) bool {
	for _, model := range repository.Models() {
		if model.ID == modelID {
			return projectAllowed(repository, value, model.ProjectID)
		}
	}
	return false
}

func agentAllowed(repository store.Repository, value auth.Principal, agentID string) bool {
	for _, agent := range repository.Agents() {
		if agent.ID == agentID {
			return projectAllowed(repository, value, agent.ProjectID)
		}
	}
	return false
}

func enforceProjectQuota(repository store.Repository, value auth.Principal) error {
	if privileged(value) || !slices.Contains(value.Roles, auth.RoleUser) {
		return nil
	}
	access, err := repository.AccessFor(value.Subject)
	if err != nil || access.Compute.MaxProjects == 0 {
		return errors.New("no project capacity has been provisioned")
	}
	owned := 0
	for _, project := range repository.Projects() {
		if project.OwnerSubject == value.Subject {
			owned++
		}
	}
	if owned >= access.Compute.MaxProjects {
		return fmt.Errorf("project quota reached (%d)", access.Compute.MaxProjects)
	}
	return nil
}

func enforceRunQuota(repository store.Repository, value auth.Principal) error {
	if privileged(value) || !slices.Contains(value.Roles, auth.RoleUser) {
		return nil
	}
	access, err := repository.AccessFor(value.Subject)
	if err != nil || access.Compute.MaxRuns == 0 {
		return errors.New("no concurrent run capacity has been provisioned")
	}
	active, allowed := 0, allowedProjectIDs(repository, value)
	for _, run := range repository.Runs() {
		if allowed[run.ProjectID] && (run.Status == "queued" || run.Status == "running") {
			active++
		}
	}
	if active >= access.Compute.MaxRuns {
		return fmt.Errorf("concurrent run quota reached (%d)", access.Compute.MaxRuns)
	}
	return nil
}

func storageAllowed(repository store.Repository, value auth.Principal, bucket string) bool {
	if privileged(value) || !slices.Contains(value.Roles, auth.RoleUser) {
		return true
	}
	access, err := repository.AccessFor(value.Subject)
	if err != nil || access.Storage.SizeGB == 0 {
		return false
	}
	if bucket == "" {
		return true
	}
	return slices.Contains(access.Storage.Buckets, bucket)
}
