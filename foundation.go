package foundation

import (
	"database/sql"
	"io/ioutil"
	"testing"

	"github.com/jmoiron/sqlx"
)

type Foundation struct {
	t            *testing.T
	currentStep  int
	stepByStep   bool
	migrator     Migrator
	db           *sqlx.DB
	interceptors map[int]Interceptor
}

type Migrator interface {
	DB() *sql.DB
	DriverName() string
	Setup() error
	MigrateToStep(step int) error
	TearDown() error
}

type Interceptor func() error

// ToDo: change T to TB?
func New(t *testing.T, migrator Migrator) *Foundation {
	if err := migrator.Setup(); err != nil {
		t.Fatalf("error setting up the migrator: %s", err)
	}

	db := sqlx.NewDb(migrator.DB(), migrator.DriverName())

	return &Foundation{
		t:           t,
		currentStep: 0,
		// if true, will run the migrator Step function once per step
		// instead of just once with the final step
		stepByStep: false,
		migrator:   migrator,
		db:         db,
	}
}

func (f *Foundation) RegisterInterceptors(interceptors map[int]Interceptor) *Foundation {
	f.interceptors = interceptors
	return f
}

func (f *Foundation) SetStepByStep(stepByStep bool) *Foundation {
	f.stepByStep = stepByStep
	return f
}

// calculateNextStep returns the next step in the chain that has an
// interceptor or the final step to migrate to
func (f *Foundation) calculateNextStep(step int) int {
	// should never happen
	if f.currentStep >= step {
		// nothing to do
		return step // ToDo: or 0? merge the two conditions
	}

	// if there are no interceptors, next step is directly the final
	// one
	if f.interceptors == nil {
		return step
	}

	i := f.currentStep
	for i < step {
		i++

		if _, ok := f.interceptors[i]; ok {
			break
		}
	}

	return i
}

func (f *Foundation) MigrateToStep(step int) *Foundation {
	if step == f.currentStep {
		// log nothing to do
		return f
	}

	if step < f.currentStep {
		f.t.Fatal("Down migrations not supported yet")
	}

	// if there are no interceptors, just migrate to the last step
	if f.interceptors == nil {
		if err := f.migrateToStep(step); err != nil {
			f.t.Fatalf("migration to step %d failed: %s", step, err)
		}

		return f
	}

	for f.currentStep < step {
		nextStep := f.calculateNextStep(step)

		if err := f.migrateToStep(nextStep); err != nil {
			f.t.Fatalf("migration to step %d failed: %s", nextStep, err)
		}

		interceptorFn, ok := f.interceptors[nextStep]
		if ok {
			if err := interceptorFn(); err != nil {
				f.t.Fatalf("interceptor function for step %d failed", nextStep)
			}
		}
	}

	return f
}

// migrateToStep executes the migrator function to migrate to a
// specific step and updates the foundation currentStep to reflect the
// result. This function doesn't take into account interceptors, that
// happens on MigrateToStep
func (f *Foundation) migrateToStep(step int) error {
	if f.stepByStep {
		for f.currentStep < step {
			if err := f.migrator.MigrateToStep(f.currentStep + 1); err != nil {
				return err
			}

			f.currentStep++
		}

		return nil
	}

	if err := f.migrator.MigrateToStep(step); err != nil {
		return err
	}

	f.currentStep = step
	return nil
}

func (f *Foundation) TearDown() {
	if err := f.migrator.TearDown(); err != nil {
		f.t.Fatalf("error tearing down migrator: %s", err)
	}
}

func (f *Foundation) DB() *sqlx.DB {
	return f.db
}

func (f *Foundation) Exec(s string) *Foundation {
	if _, err := f.DB().Exec(s); err != nil {
		f.t.Fatalf("failed to run %s: %s", s, err)
	}

	return f
}

func (f *Foundation) ExecFile(filePath string) *Foundation {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		f.t.Fatalf("failed to read file %s: %s", filePath, err)
	}

	return f.Exec(string(b))
}
