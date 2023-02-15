package mysql

import (
	"database/sql"

	"moduls/pkg/models"
)

// code

// Define a SnippetModel type which wraps a sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

func (m *SnippetModel) Insert(title, content, expires string) (int, error) {
	stmt := `INSERT INTO snippets (title, content, created, expires)
    VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	// Write the SQL statement we want to execute. Again, I've split it over tw
	// lines for readability.
	stmt := `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() AND id = ?`
	// Use the QueryRow() method on the connection pool to execute our
	// SQL statement, passing in the untrusted id variable as the value for the
	// placeholder parameter. This returns a pointer to a sql.Row object which
	// holds the result from the database.
	row := m.DB.QueryRow(stmt, id)
	// Initialize a pointer to a new zeroed Snippet struct.
	s := &models.Snippet{}
	// Use row.Scan() to copy the values from each field in sql.Row to the
	// corresponding field in the Snippet struct. Notice that the arguments
	// to row.Scan are *pointers* to the place you want to copy the data into,
	// and the number of arguments must be exactly the same as the number of
	// columns returned by your statement. If the query returns no rows, then
	// row.Scan() will return a sql.ErrNoRows error. We check for that and retu
	// our own models.ErrNoRecord error instead of a Snippet object.
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err == sql.ErrNoRows {
		return nil, models.ErrNoRecord
	} else if err != nil {
		return nil, err
	}
	// If everything went OK then return the Snippet object.
	return s, nil
}

func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return nil, nil
}
