package storage

import (
	"database/sql"
	"time"

	_ "github.com/glebarez/go-sqlite"

	"daily-tracker/internal/model"
)

type SqliteDB struct {
	db *sql.DB
}

func New(path string) (*SqliteDB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &SqliteDB{db: db}
	return s, s.migrate()
}

func (db *SqliteDB) SaveHabit(h model.Habit) (int64, error) {
	var goalVal sql.NullFloat64
	if h.Goal != nil {
		goalVal = sql.NullFloat64{Float64: *h.Goal, Valid: true}
	}

	var endDateVal sql.NullString
	if h.EndDate != nil {
		endDateVal = sql.NullString{String: h.EndDate.Format("2006-01-02"), Valid: true}
	}

	res, err := db.db.Exec(
		`INSERT INTO habits (name, type, goal, start_date, end_date, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		h.Name,
		h.Type,
		goalVal,
		h.StartDate.Format("2006-01-02"),
		endDateVal,
		h.CreatedAt.Format("2006-01-02"),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *SqliteDB) GetHabits(date time.Time) ([]model.Habit, error) {
	startMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	nextMonth := startMonth.AddDate(0, 1, 0)
	rows, err := db.db.Query(
		`
                SELECT id, name, type, goal, start_date, end_date, created_at
                FROM habits
                WHERE start_date < ?
                  AND (end_date IS NULL OR end_date >= ?)
                ORDER BY id
        `,
		nextMonth.Format("2006-01-02"),
		date.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	// defer - вызывает участок кода, после завершения функции, очень удобно
	// defer rows.Close()
	defer func() {
		if err := rows.Close(); err != nil {
			_ = err
		}
	}()

	var habits []model.Habit

	for rows.Next() {
		var habit model.Habit
		var goal sql.NullFloat64
		var endDate sql.NullString
		var startDate, createdAt string

		err := rows.Scan(
			&habit.ID,
			&habit.Name,
			&habit.Type,
			&goal,
			&startDate,
			&endDate,
			&createdAt,
		)
		if err != nil {
			return nil, err
		}

		if goal.Valid {
			habit.Goal = &goal.Float64
		}

		if endDate.Valid {
			t, _ := time.Parse("2006-01-02", endDate.String)
			habit.EndDate = &t
		}

		habit.StartDate, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, err
		}
		habit.CreatedAt, err = time.Parse("2006-01-02", createdAt)
		if err != nil {
			return nil, err
		}
		habits = append(habits, habit)
	}
	return habits, rows.Err()
}

func (db *SqliteDB) GetHabitByID(id int64) (model.Habit, error) {
	var habit model.Habit
	var goal sql.NullFloat64
	var endDate sql.NullString
	var startDate, createdAt string

	rows, err := db.db.Query(`
	SELECT id, name, type, goal, start_date, end_date, created_at
	FROM habits
	WHERE id = ?
	`, id)
	if err != nil {
		return model.Habit{}, err
	}
	for rows.Next() {
		err := rows.Scan(
			&habit.ID,
			&habit.Name,
			&habit.Type,
			&goal,
			&startDate,
			&endDate,
			&createdAt,
		)
		if err != nil {
			return model.Habit{}, err
		}
		if goal.Valid {
			habit.Goal = &goal.Float64
		}
		if endDate.Valid {
			t, err := time.Parse("2006-01-02", endDate.String)
			if err != nil {
				return model.Habit{}, err
			}
			habit.EndDate = &t
		}

		habit.StartDate, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return model.Habit{}, err
		}

		habit.CreatedAt, err = time.Parse("2006-01-02", createdAt)
		if err != nil {
			return model.Habit{}, err
		}

	}
	return habit, nil
}

func (db *SqliteDB) UpdateHabit(h model.Habit) (model.Habit, error) {
	var goalVal sql.NullFloat64
	if h.Goal != nil {
		goalVal = sql.NullFloat64{Float64: *h.Goal, Valid: true}
	}

	var endDateVal sql.NullString
	if h.EndDate != nil {
		endDateVal = sql.NullString{String: h.EndDate.Format("2006-01-02"), Valid: true}
	}

	_, err := db.db.Exec(
		`UPDATE habits SET name = ?, type = ?, goal = ?, start_date = ?, end_date = ?, created_at = ? WHERE id = ?`,
		h.Name,
		h.Type,
		goalVal,
		h.StartDate.Format("2006-01-02"),
		endDateVal,
		h.CreatedAt.Format("2006-01-02"),
		h.ID,
	)
	if err != nil {
		return model.Habit{}, err
	}

	var habit model.Habit
	var goal sql.NullFloat64
	var endDate sql.NullString
	var startDate, createdAt string

	row := db.db.QueryRow(
		`SELECT id, name, type, goal, start_date, end_date, created_at FROM habits WHERE id = ?`,
		h.ID,
	)
	err = row.Scan(
		&habit.ID,
		&habit.Name,
		&habit.Type,
		&goal,
		&startDate,
		&endDate,
		&createdAt,
	)
	if err != nil {
		return model.Habit{}, err
	}

	if goal.Valid {
		habit.Goal = &goal.Float64
	}

	if endDate.Valid {
		t, _ := time.Parse("2006-01-02", endDate.String)
		habit.EndDate = &t
	}

	habit.StartDate, _ = time.Parse("2006-01-02", startDate)
	habit.CreatedAt, _ = time.Parse("2006-01-02", createdAt)

	return habit, nil
}

func (db *SqliteDB) HasHabitByDate(date time.Time) (bool, error) {
	monthStart := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	nextMonth := monthStart.AddDate(0, 1, 0)
	var exist bool

	row := db.db.QueryRow(
		`SELECT EXISTS(
		SELECT	1
		FROM habits
		WHERE start_date < ? AND (end_date IS NULL OR end_date >= ?)
		LIMIT 1)
		`,
		nextMonth.Format("2006-01-02"), monthStart.Format("2006-01-02"),
	)

	err := row.Scan(&exist)
	return exist, err
}

func (db *SqliteDB) DeleteHabit(id int64) error {
	_, err := db.db.Exec(`DELETE FROM habits WHERE id = ?`, id)
	return err
}

func (db *SqliteDB) HabitEndDate(id int64, date time.Time) error {
	_, err := db.db.Exec(
		`UPDATE habits SET end_date = ? WHERE id = ?`,
		date.Format("2006-01-02"),
		id,
	)
	return err
}

func (db *SqliteDB) SaveEntry(entry model.Entry) error {
	_, err := db.db.Exec(
		`INSERT INTO entries (habit_id, date, value) VALUES (?, ?, ?)`,
		entry.HabitID,
		entry.Date.Format("2006-01-02"),
		entry.Value,
	)
	return err
}

func (db *SqliteDB) GetEntries(date time.Time) ([]model.Entry, error) {
	startDate := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	nextMonth := startDate.AddDate(0, 1, 0)
	rows, err := db.db.Query(
		`
	SELECT id, habit_id, date, value
	FROM entries
	WHERE date >= ? AND date < ?
	`,
		startDate.Format("2006-01-02"),
		nextMonth.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []model.Entry

	for rows.Next() {
		var e model.Entry
		var date string

		err := rows.Scan(&e.ID, &e.HabitID, &date, &e.Value)
		if err != nil {
			return nil, err
		}

		e.Date, _ = time.Parse("2006-01-02", date)
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

func (db *SqliteDB) UpdateEntry(id int64, value float64) (model.Entry, error) {
	_, err := db.db.Exec(`UPDATE entries SET value = ? WHERE id = ?`, value, id)
	if err != nil {
		return model.Entry{}, err
	}

	var e model.Entry
	var date string

	row := db.db.QueryRow(`SELECT id, habit_id, date, value FROM entries WHERE id = ?`, id)
	err = row.Scan(&e.ID, &e.HabitID, &date, &e.Value)
	if err != nil {
		return model.Entry{}, err
	}

	e.Date, _ = time.Parse("2006-01-02", date)
	return e, nil
}

func (db *SqliteDB) DeleteEntry(id int64) error {
	_, err := db.db.Exec(`DELETE FROM entries WHERE id = ?`, id)
	return err
}

func (db *SqliteDB) migrate() error {
	_, err := db.db.Exec(`
		CREATE TABLE IF NOT EXISTS habits (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT NOT NULL,
			type       TEXT NOT NULL,
			goal       REAL,
			start_date TEXT NOT NULL,
			end_date   TEXT,
			created_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS entries (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			habit_id INTEGER NOT NULL REFERENCES habits(id),
			date     TEXT NOT NULL,
			value    REAL NOT NULL
		);
	`)
	return err
}

var _ Storage = (*SqliteDB)(nil)
