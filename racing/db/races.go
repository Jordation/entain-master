package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

const (
	RACE_STATUS_OPEN   = "OPEN"
	RACE_STATUS_CLOSED = "CLOSED"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)

	// Get will return a single race matching the ID of the request
	Get(raceId int64) (*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

// Linked function for RacingService.GetRace
func (r *racesRepo) Get(raceId int64) (*racing.Race, error) {
	query := fmt.Sprintf(getRaceQueries()[racesGet], raceId)

	row, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}

	res, err := r.scanRaces(row)
	if err != nil {
		return nil, err
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("get race did not return 1 race")
	}
	return res[0], nil
}

// Linked function for RacingService.ListRaces
func (r *racesRepo) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}

	if filter.OnlyVisible {
		clauses = append(clauses, "visible = ?")
		args = append(args, 1)
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	// ORDER BY applied after query is joined with clauses
	if filter.ASTOrderBy != nil {
		AstOrder := filter.GetASTOrderBy()

		// Validate the parsed value
		if AstOrder != ("DESC") && AstOrder != ("ASC") {
			log.Info("[RACING]: invalid order by value")
		} else {
			query += fmt.Sprintf(" ORDER BY advertised_start_time %v", AstOrder)
		}
	}

	return query, args
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race
	currentTime := time.Now()

	for rows.Next() {
		var (
			advertisedStart time.Time
			race            racing.Race
		)

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		// derives race.Status from current time compared against scanned advertisedStart time
		if currentTime.Before(advertisedStart) {
			race.Status = RACE_STATUS_OPEN
		} else {
			race.Status = RACE_STATUS_CLOSED
		}

		race.AdvertisedStartTime = ts

		races = append(races, &race)
	}

	return races, nil
}
