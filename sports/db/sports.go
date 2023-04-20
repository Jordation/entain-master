package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"git.neds.sh/matty/entain/sports/proto/sports_pb"
	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"
)

const (
	SPORT_STATUS_OPEN   = "OPEN"
	SPORT_STATUS_CLOSED = "CLOSED"
)

// SportsRepo provides repository access to sports.
type SportsRepo interface {
	// Init will initialise our sports repository.
	Init() error

	// List will return a list of sports.
	List(filter *sports_pb.ListSportsRequestFilter) ([]*sports_pb.Sport, error)

	// Get will return a single sport matching the ID of the request
	Get(sportId int64) (*sports_pb.Sport, error)
}

type sportsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewSportsRepo creates a new sports repository.
func NewSportsRepo(db *sql.DB) SportsRepo {
	return &sportsRepo{db: db}
}

// Init prepares the sport repository dummy data.
func (s *sportsRepo) Init() error {
	var err error

	s.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy sports_pb.
		err = s.seed()
	})

	return err
}

// Didn't want to change module go version but should be generic
// Functions to check if an element is within a slice
func contains(e string, s []string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// Linked function for SportsService.GetSport
func (s *sportsRepo) Get(sportId int64) (*sports_pb.Sport, error) {
	query := fmt.Sprintf(getSportQueries()[sportsGet], sportId)

	row, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	res, err := s.scanSports(row)
	if err != nil {
		return nil, err
	}

	return res[0], nil
}

// Linked function for SportsService.ListSports
func (s *sportsRepo) List(filter *sports_pb.ListSportsRequestFilter) ([]*sports_pb.Sport, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	if filter.OrderBy != nil {
		cols, err := s.getSportDbColumns()
		if err != nil {
			return nil, err
		}
		if !s.validateOrderByFilter(filter.OrderBy, cols) {
			return nil, fmt.Errorf("[LIST SPORTS]: invalid order by filter")
		}
	}

	query = getSportQueries()[sportsList]

	query, args = s.applyFilter(query, filter)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return s.scanSports(rows)
}

// Validates that the field provided matches a column in the database
func (s *sportsRepo) validateOrderByFilter(f *sports_pb.SportsOrderByFilter, cols []string) bool {
	if contains(f.OrderByField, cols) &&
		((f.OrderByDirection == "ASC") || (f.OrderByDirection == "DESC")) {
		return true
	}
	return false
}

// Returns ORDER BY clause for concatenation
func getOrderByClause(filter *sports_pb.SportsOrderByFilter) string {
	return fmt.Sprintf(
		getSportQueries()[sportsOrderby],
		filter.OrderByField,
		filter.OrderByDirection,
	)
}

func (s *sportsRepo) applyFilter(query string, filter *sports_pb.ListSportsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.EventIds) > 0 {
		clauses = append(clauses, "event_id IN ("+strings.Repeat("?,", len(filter.EventIds)-1)+"?)")

		for _, eventID := range filter.EventIds {
			args = append(args, eventID)
		}
	}

	if filter.OnlyVisible {
		clauses = append(clauses, "visible = ?")
		args = append(args, 1)
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	if filter.OrderBy != nil {
		query += getOrderByClause(filter.OrderBy)
	}

	return query, args
}

func (m *sportsRepo) scanSports(
	rows *sql.Rows,
) ([]*sports_pb.Sport, error) {
	var (
		sports      []*sports_pb.Sport
		currentTime = time.Now()
	)

	for rows.Next() {
		var (
			advertisedStart time.Time
			sport           sports_pb.Sport
		)

		if err := rows.Scan(&sport.Id, &sport.EventId, &sport.Category, &sport.Team_1, &sport.Team_2, &sport.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		// derives sport.Status from current time compared against scanned advertisedStart time
		if currentTime.Before(advertisedStart) {
			sport.Status = SPORT_STATUS_OPEN
		} else {
			sport.Status = SPORT_STATUS_CLOSED
		}

		sport.AdvertisedStartTime = ts

		sports = append(sports, &sport)
	}

	return sports, nil
}

// Gets a list of current columns from Sports DB table
func (s *sportsRepo) getSportDbColumns() ([]string, error) {
	var columnsNames []string

	rows, err := s.db.Query(getSportQueries()[sportsListDbColumns])
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			return nil, err
		}

		columnsNames = append(columnsNames, colName)
	}

	return columnsNames, nil
}
