package database

import (
	"context"
	"github.com/rs/zerolog/log"
	"os"
)

type Migration struct {
	id        int
	migration string
}

func (db *Database) Migrate(ctx context.Context) error {
	dir := "./migrations/"
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Debug().Err(err).Msg("Nothing to executeMigration")
		return err
	}

	runMigrations, err := db.getRunMigrations(ctx)
	if err != nil {
		return err
	}

	hasRun := 0
	for _, file := range files {
		migrationName := file.Name()[:len(file.Name())-4]

		_, ok := runMigrations[migrationName]
		if ok {
			continue
		}

		sql, err := os.ReadFile(dir + file.Name())
		if err != nil {
			log.Warn().Err(err).Msgf("Cannot open file [%s]", file.Name())
			return err
		}

		log.Debug().Msgf("[%s] Migrating...", migrationName)
		if err := db.executeMigration(ctx, migrationName, string(sql)); err != nil {
			log.Warn().Err(err).Msgf("[%s] Migration failed", migrationName)
			return err
		}
		hasRun++
		log.Debug().Msgf("[%s] Migrated!", migrationName)
	}

	if hasRun > 0 {
		log.Info().Msg("FindAll migrations have run successfully!")
	} else {
		log.Info().Msg("Nothing to executeMigration")
	}
	return nil
}

func (db *Database) executeMigration(ctx context.Context, migrationName, migrationSQL string) error {
	if _, err := db.Exec(ctx, migrationSQL); err != nil {
		return err
	}

	saveMigrationSQL := `INSERT INTO "public"."migration" (migration) VALUES ($1);`
	if _, err := db.Exec(ctx, saveMigrationSQL, migrationName); err != nil {
		return err
	}

	return nil
}

func (db *Database) createMigrationTable(ctx context.Context) error {
	sql := `
		CREATE TABLE IF NOT EXISTS "migration" (
		    id			SERIAL PRIMARY KEY,
		    migration	VARCHAR(255) NOT NULL UNIQUE
		)
	`

	if _, err := db.Exec(ctx, sql); err != nil {
		return err
	}

	return nil
}

func (db *Database) getRunMigrations(ctx context.Context) (map[string]Migration, error) {
	if err := db.createMigrationTable(ctx); err != nil {
		return map[string]Migration{}, err
	}

	runMigrations := make(map[string]Migration, 0)

	rows, err := db.Query(ctx, `SELECT id, migration FROM "public"."migration";`)
	if err != nil {
		return map[string]Migration{}, err
	}

	for rows.Next() {
		runMigration := new(Migration)

		if err = rows.Scan(&runMigration.id, &runMigration.migration); err != nil {
			return map[string]Migration{}, err
		}

		runMigrations[runMigration.migration] = *runMigration
	}

	return runMigrations, nil
}
