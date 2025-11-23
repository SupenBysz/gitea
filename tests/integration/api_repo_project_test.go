// Copyright 2025 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"testing"

	auth_model "code.gitea.io/gitea/models/auth"
	project_model "code.gitea.io/gitea/models/project"
	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/tests"

	"github.com/stretchr/testify/assert"
)

func TestAPIListProjects(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	project := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: 1})
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: project.RepoID})
	repoOwner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

	session := loginUser(t, repoOwner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)

	req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects", repoOwner.Name, repo.Name)).
		AddTokenAuth(token)
	resp := session.MakeRequest(t, req, http.StatusOK)

	var projects []api.Project
	DecodeJSON(t, resp, &projects)

	assert.GreaterOrEqual(t, len(projects), 1)
}

func TestAPIGetProject(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	project := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: 1})
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: project.RepoID})
	repoOwner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

	session := loginUser(t, repoOwner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)

	req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d", repoOwner.Name, repo.Name, project.ID)).
		AddTokenAuth(token)
	resp := session.MakeRequest(t, req, http.StatusOK)

	apiProject := new(api.Project)
	DecodeJSON(t, resp, &apiProject)

	assert.Equal(t, project.ID, apiProject.ID)
	assert.Equal(t, project.Title, apiProject.Title)
}

func TestAPICreateProject(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
	repoOwner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

	session := loginUser(t, repoOwner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	opts := api.CreateProjectOption{
		Title: "Test Project API",
	}

	req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects", repoOwner.Name, repo.Name), &opts).
		AddTokenAuth(token)
	resp := session.MakeRequest(t, req, http.StatusCreated)

	apiProject := new(api.Project)
	DecodeJSON(t, resp, &apiProject)

	assert.Equal(t, "Test Project API", apiProject.Title)

	// Verify it was created in the database
	dbProject := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: apiProject.ID})
	assert.Equal(t, "Test Project API", dbProject.Title)
}

func TestAPIListProjectColumns(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	project := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: 1})
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: project.RepoID})
	repoOwner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

	session := loginUser(t, repoOwner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeReadRepository)

	req := NewRequest(t, "GET", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns", repoOwner.Name, repo.Name, project.ID)).
		AddTokenAuth(token)
	resp := session.MakeRequest(t, req, http.StatusOK)

	var columns []api.ProjectColumn
	DecodeJSON(t, resp, &columns)

	assert.GreaterOrEqual(t, len(columns), 0)
}

func TestAPICreateProjectColumn(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	project := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: 1})
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: project.RepoID})
	repoOwner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

	session := loginUser(t, repoOwner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	opts := api.CreateProjectColumnOption{
		Title: "Test Column API",
	}

	req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d/columns", repoOwner.Name, repo.Name, project.ID), &opts).
		AddTokenAuth(token)
	resp := session.MakeRequest(t, req, http.StatusCreated)

	apiColumn := new(api.ProjectColumn)
	DecodeJSON(t, resp, &apiColumn)

	assert.Equal(t, "Test Column API", apiColumn.Title)

	// Verify it was created in the database
	dbColumn := unittest.AssertExistsAndLoadBean(t, &project_model.Column{ID: apiColumn.ID})
	assert.Equal(t, "Test Column API", dbColumn.Title)
}

func TestAPIDeleteProject(t *testing.T) {
	defer tests.PrepareTestEnv(t)()

	project := unittest.AssertExistsAndLoadBean(t, &project_model.Project{ID: 6}) // Use a project that can be safely deleted
	repo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: project.RepoID})
	repoOwner := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: repo.OwnerID})

	session := loginUser(t, repoOwner.Name)
	token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

	req := NewRequest(t, "DELETE", fmt.Sprintf("/api/v1/repos/%s/%s/projects/%d", repoOwner.Name, repo.Name, project.ID)).
		AddTokenAuth(token)
	session.MakeRequest(t, req, http.StatusNoContent)

	// Verify it was deleted from the database
	unittest.AssertNotExistsBean(t, &project_model.Project{ID: project.ID})
}
