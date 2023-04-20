package db

import (
	"time"

	"syreclabs.com/go/faker"
)

func (s *sportsRepo) seed() error {
	statement, err := s.db.Prepare(`CREATE TABLE IF NOT EXISTS sports (id INTEGER PRIMARY KEY, event_id INTEGER, category TEXT, team_1 TEXT, team_2 TEXT, visible INTEGERY, advertised_start_time DATETIME)`)
	if err == nil {
		_, err = statement.Exec()
	}

	for i := 1; i <= 100; i++ {
		statement, err = s.db.Prepare(`INSERT OR IGNORE INTO sports(id, event_id, category, team_1, team_2, visible, advertised_start_time) VALUES (?,?,?,?,?,?,?)`)
		if err == nil {
			_, err = statement.Exec(
				i,
				faker.Number().Between(1, 20),
				faker.Lorem().Word(),
				faker.Team().Name(),
				faker.Team().Name(),
				faker.Number().Between(0, 1),
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
			)
		}
	}

	return err
}
