package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "github.com/crusttech/crust/sam/db/mysql"
	"github.com/pkg/errors"
	"github.com/rakyll/statik/fs"
	"github.com/titpetric/factory"
)

//go:generate statik -p mysql -Z -f -src=schema/mysql

func statements(contents []byte, err error) ([]string, error) {
	if err != nil {
		return []string{}, err
	}
	return regexp.MustCompilePOSIX(";$").Split(string(contents), -1), nil
}

type migration struct {
	Project        string
	Filename       string
	StatementIndex int `db:"statement_index"`
	Status         string
}

func Migrate(db *factory.DB) error {
	statikFS, err := fs.New()
	if err != nil {
		return errors.Wrap(err, "Error creating statik filesystem")
	}

	var files []string

	if err := fs.Walk(statikFS, "/", func(filename string, info os.FileInfo, err error) error {
		matched, err := filepath.Match("/*.up.sql", filename)
		if matched {
			files = append(files, filename)
		}
		return err
	}); err != nil {
		return errors.Wrap(err, "Error when listing files for migrations")
	}

	sort.Strings(files)

	if len(files) == 0 {
		return errors.New("No files encoded for migration, need at least one SQL file")
	}

	migrate := func(filename string, useLog bool) error {
		status := migration{
			Project:  "sam",
			Filename: filename,
		}
		if useLog {
			if err := db.Get(&status, "select * from migrations where project=? and filename=?", status.Project, status.Filename); err != nil {
				return err
			}
			if status.Status == "ok" {
				return nil
			}
		}

		up := func() error {
			stmts, err := statements(fs.ReadFile(statikFS, filename))
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Error reading migration %s", filename))
			}

			log.Println("Running migration for", filename)
			for idx, query := range stmts {
				if strings.TrimSpace(query) != "" && idx >= status.StatementIndex {
					status.StatementIndex = idx
					if _, err := db.Exec(query); err != nil {
						return err
					}
				}
			}
			log.Println("Migration done OK")
			status.Status = "ok"
			return nil
		}

		if err := db.Transaction(up); err != nil {
			status.Status = err.Error()
			if useLog {
				db.Replace("migrations", status)
			}
			return err
		}
		return nil
	}

	if err := migrate("/migrations.sql", false); err != nil {
		return err
	}

	for _, filename := range files {
		if err := migrate(filename, true); err != nil {
			return err
		}
	}

	return nil
}