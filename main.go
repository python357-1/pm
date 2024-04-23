package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type ProjectStep struct {
	Id          string
	ProjectId   string
	StepNumber  int
	Description string
}

func (ps *ProjectStep) ToString() string {
	//fmt.Sprintf("ProjectStep {Id: ()\n\tProjectId: ()\n\tStepNumber: ()\n\tDescription: ()\n\t}")
	return fmt.Sprintf("%#v", ps)
}

func ProjectStepsToString(ps []ProjectStep) string {
	sb := strings.Builder{}
	for i, p := range ps {
		sb.WriteString(fmt.Sprint(i) + ": " + p.ToString() + "\n")
	}

	return sb.String()
}

type Project struct {
	Name        string
	Description string
	Id          string
	Steps       []ProjectStep
	WithEditRow bool
}

func (p *Project) ToString() string {
	return "Project {\n\tName: " + p.Name + "\n\tDescription: " + p.Description + "\n\tId: " + p.Id + "\n\tCompletionSteps: []ProjectStep{" + ProjectStepsToString(p.Steps) + "}\n}"
}

type ProjectViewModel struct {
	Projects        []Project
	SelectedProject Project
}

func panicIfErr(e error) {
	if e != nil {
		panic(e)
	}
}

func parse(x string) (int, error) {
	return strconv.Atoi(x)
}

func add(x, y int) string {
	return strconv.Itoa(x + y)

}

func main() {
	mux := http.NewServeMux()
	repo := NewRepository()
	funcMap := template.FuncMap{
		"add":   add,
		"parse": parse,
	}

	templates := template.Must(template.New("stupidfuckingbitch").Funcs(funcMap).ParseFiles("main.html", "projects-list.html", "projects-description.html", "projects-steps-table.html"))
	mux.HandleFunc("POST /projects/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Matched POST /projects")
		p := Project{
			Id:          uuid.NewString(),
			Name:        r.FormValue("project-name"),
			Description: r.FormValue("project-description"),
		}
		repo.AddProject(p)
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	})

	mux.HandleFunc("DELETE /projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Matched DELETE /projects/{id}")
		projectId := r.PathValue("id")
		repo.RemoveProject(projectId)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Matched GET /")
		selectedProject, err := repo.GetProjectByIndex(0)
		fmt.Println("Selected Project is " + selectedProject.ToString())
		if err != nil {
			panic(err)
		}
		templates.ExecuteTemplate(w, "base", ProjectViewModel{Projects: repo.GetAllProjects(), SelectedProject: selectedProject})
	})

	mux.HandleFunc("GET /projects/{selectedId}", func(w http.ResponseWriter, r *http.Request) {
		selectedId := r.PathValue("selectedId")
		fmt.Println("Matched GET /{selectedId}")
		selectedProj, err := repo.GetProjectById(selectedId)
		if err != nil {
			http.NotFound(w, r)
		}
		projects := repo.GetAllProjects()

		templates.ExecuteTemplate(w, "base", ProjectViewModel{Projects: projects, SelectedProject: selectedProj})
	})

	mux.HandleFunc("GET /projects/{id}/description", func(w http.ResponseWriter, r *http.Request) {
		projectId := r.PathValue("id")
		fmt.Println("Matched GET /projects/" + projectId + "/description")
		proj, err := repo.GetProjectById(projectId)
		if err != nil {
			panic(err)
		}

		fmt.Fprint(w, proj.Description)
	})

	mux.HandleFunc("PUT /projects/{id}/description", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Matched PUT /projects/{id}/description")
		projDescription := r.FormValue("project-description")
		projId := r.PathValue("id")
		repo.SetProjectDescription(projId, projDescription)
		fmt.Fprint(w, http.StatusOK)
	})

	mux.HandleFunc("GET /project/{id}/steps/", func(w http.ResponseWriter, r *http.Request) {
		projectId := r.PathValue("id")
		fmt.Println("Matched GET /project/" + projectId + "/steps/")
		wer := r.FormValue("WithEditRow")
		WithEditRow, err := strconv.ParseBool(wer)
		fmt.Println("WithEditRow: " + wer)
		if err != nil {
			WithEditRow = false
		}
		proj, err := repo.GetProjectById(projectId)
		panicIfErr(err)

		proj.WithEditRow = WithEditRow

		newerr := templates.ExecuteTemplate(w, "projectsStepsTable", ProjectViewModel{SelectedProject: proj})
		if newerr != nil {
			fmt.Fprint(w, "<div style=\"color: red\">"+newerr.Error()+"</div>")
		}
	})

	mux.HandleFunc("POST /project/{id}/steps/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Matched POST /projectsteps/")
		projectId := r.PathValue("id")
		sn := r.FormValue("stepNumber")
		Description := r.FormValue("Description")
		stepNumber, err := strconv.Atoi(sn)
		panicIfErr(err)
		step := ProjectStep{Id: uuid.NewString(), ProjectId: projectId, StepNumber: stepNumber, Description: Description}
		repo.AddStepToProject(projectId, step)
		http.Redirect(w, r, "/projects/"+projectId, http.StatusSeeOther)
	})

	port := ":8080"
	fmt.Println("Listening on http://localhost" + port)
	http.ListenAndServe(port, mux)
}