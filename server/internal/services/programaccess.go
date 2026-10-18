package services

import (
	"context"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
)

// projectLeadLookup tells whether a person leads one of a program's projects.
type projectLeadLookup interface {
	LeadsProjectIn(ctx context.Context, programID, personID string) (bool, error)
	OnBaselineIn(ctx context.Context, programID, personID string) (bool, error)
}

// requireProgramReader lets a planner read any program, and a manager a program they lead or
// lead one of its projects. A specialist reads those too, and a program where they have baseline
// hours. Anyone else gets a 403 error.
func requireProgramReader(ctx context.Context, authz AuthzService, projects projectLeadLookup, program *models.Program, op string) error {
	reader, err := authz.ProgramReader(ctx)
	if err != nil {
		return err
	}
	if reader.All {
		return nil
	}
	if reader.LeadID != "" {
		if isPerson(program.LeadID, reader.LeadID) {
			return nil
		}
		led, err := projects.LeadsProjectIn(ctx, program.ID, reader.LeadID)
		if err != nil {
			return errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceProgram)
		}
		if led {
			return nil
		}
		if reader.OnBaseline {
			worked, err := projects.OnBaselineIn(ctx, program.ID, reader.LeadID)
			if err != nil {
				return errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceProgram)
			}
			if worked {
				return nil
			}
		}
	}
	return errors.ForbiddenError(constants.MsgInsufficientPermissions, nil).WithOperation(op).WithResource(resourceProgram)
}

// readerLeadID is nil for a reader of every program, else the manager whose programs a list keeps.
func readerLeadID(reader ProgramReader) *string {
	if reader.All {
		return nil
	}
	return &reader.LeadID
}

// emptyPage is a page without rows, for a reader who may see none.
func emptyPage[T any](page *dtos.PageableRequest) *dtos.DataResponse[T] {
	return &dtos.DataResponse[T]{Data: []T{}, Pageable: &dtos.Pageable{Page: page.Page, PageSize: page.PageSize}}
}

// programOf is a project's program, or a bare one with its id when it was not loaded.
func programOf(project *models.Project) *models.Program {
	if project.Program != nil {
		return project.Program
	}
	return &models.Program{BaseModel: models.BaseModel{ID: project.ProgramID}}
}

// readablePrograms tells which of these projects' programs the reader may read, from the
// projects themselves: a manager reads a program they lead or lead one of its projects.
func readablePrograms(projects []models.Project, reader ProgramReader) func(programID string) bool {
	readable := map[string]bool{}
	for i := range projects {
		project := &projects[i]
		if reader.LeadID != "" && (isPerson(project.LeadID, reader.LeadID) || isPerson(programOf(project).LeadID, reader.LeadID)) {
			readable[project.ProgramID] = true
		}
	}
	return func(programID string) bool { return reader.All || readable[programID] }
}

// isPerson reports whether an optional person id is this person.
func isPerson(id *string, personID string) bool {
	return id != nil && *id == personID
}
