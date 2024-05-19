package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type Repository struct {
	projects     []Project
	projectSteps []ProjectStep
}

func NewRepository() Repository {
	r := Repository{}
	r.Init()
	return r
}

func (r *Repository) GetAllProjects() []Project {
	return r.projects
}

func (r *Repository) GetProjectSteps(projectId string) []ProjectStep {
	items := []ProjectStep{}

	for _, ps := range r.projectSteps {
		if ps.ProjectId == projectId {
			items = append(items, ps)
		}
	}

	return items
}

func (r *Repository) Init() {
	p := Project{"TestProj1", "This is a description for a thing that is being described", uuid.NewString(), []ProjectStep{}, false}
	ps := []ProjectStep{
		{uuid.NewString(), p.Id, 1, "Step Description"},
		{uuid.NewString(), p.Id, 2, "Step Description 2"},
	}
	p.Steps = ps
	r.projects = append(r.projects, p)
}

func (r *Repository) AddProject(p Project) {
	r.projects = append(r.projects, p)
}

func (r *Repository) RemoveProject(id string) {
	if len(r.projects) == 1 {
		return
	}
	tempProjects := []Project{}

	for _, p := range r.projects {
		if p.Id != id {
			tempProjects = append(tempProjects, p)
		}
	}

	r.projects = tempProjects
}

func (r *Repository) GetProjectById(id string) (Project, error) {
	for i, p := range r.projects {
		if p.Id == id {
			return r.projects[i], nil
		}
	}

	return Project{}, errors.New("could not find project")
}

func (r *Repository) GetProjectByIndex(idx int) (Project, error) {
	if idx >= len(r.projects) {
		return Project{}, errors.New("index too large")
	}

	return r.projects[idx], nil
}

func (r *Repository) SetProjectDescription(projectId, desc string) {
	for i, p := range r.projects {
		if p.Id == projectId {
			r.projects[i].Description = desc
		}
	}
}

func (r *Repository) AddStepToProject(projectId string, step ProjectStep) (bool, error) {
	_, err := r.GetProjectById(projectId)
	if err != nil {
		return false, err
	}

	for i, p := range r.projects {
		if p.Id == projectId {
			r.projects[i].Steps = append(r.projects[i].Steps, step)
		}
	}

	return true, nil
}

func (r *Repository) ImportJson(jsonData string) {
	var tempProjects []Project
	err := json.Unmarshal([]byte(jsonData), &tempProjects)
	panicIfErr(err)

	fmt.Println(tempProjects)
	for _, p := range tempProjects {
		_, err := r.GetProjectById(p.Id)
		if err == nil {
			continue
		}
		r.AddProject(p)
		for _, ps := range p.Steps {
			r.AddStepToProject(p.Id, ps)
		}
	}
}
