package foundation

import "database/sql"

type SampleMigrator struct{}

func (sm *SampleMigrator) DB() *sql.DB {
	return nil
}

func (sm *SampleMigrator) DriverName() string {
	return ""
}

func (sm *SampleMigrator) Setup() error {
	return nil
}

func (sm *SampleMigrator) MigrateToStep(step int) error {
	return nil
}

func (sm *SampleMigrator) TearDown() error {
	return nil
}

var _ Migrator = &SampleMigrator{}
