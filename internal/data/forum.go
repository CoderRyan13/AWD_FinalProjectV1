// Filename: internal/data/forum.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"AWD_FinalProject.ryanarmstrong.net/internal/validator"
	"github.com/lib/pq"
)

type Forum struct {
	ID         int64     `json:"id"` // Struct tags
	User_ID    int64     `json:"user_id"`
	CreatedAt  time.Time `json:"created_at"`
	Topic      string    `json:"topic"`
	Discussion string    `json:"discussion"`
	Comments   []string  `json:"comments"`
	Version    int32     `json:"version"`
}

func ValidateForum(v *validator.Validator, forum *Forum) {
	// Use the Check() method to execute our validation checks
	v.Check(forum.Topic != "", "topic", "must be provided")
	v.Check(len(forum.Topic) <= 200, "topic", "must not be more than 200 bytes long")

	v.Check(forum.Discussion != "", "discussion", "must be provided")

	//v.Check(forum.Discussion != nil, "discussion", "must be provided")
	//v.Check(len(forum.Discussion) >= 1, "discussion", "must contain at least 1 entry")
	v.Check(validator.Unique(forum.Comments), "comments", "must not contain duplicate entries")
}

// Define a ForumModel which wraps a sql.DB connection pool
type ForumModel struct {
	DB *sql.DB
}

// Insert() allows us to create a new Forum
func (m ForumModel) Insert(forum *Forum) error {
	query := `
		INSERT INTO forums (topic, discussion)
		VALUES ($1, $2)
		RETURNING id, user_id, created_at, version, comments
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Collect the data fields into a slice
	args := []interface{}{
		forum.Topic, forum.Discussion,
	}
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&forum.ID, &forum.User_ID, &forum.CreatedAt, &forum.Version, pq.Array(&forum.Comments))
}

// Get() allows us to recieve a specific Forum
func (m ForumModel) Get(id int64) (*Forum, error) {
	// Ensure that there is a valid id
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create the query
	query := `
		SELECT id, user_id, created_at, topic, discussion, version, comments
		FROM forums
		WHERE id = $1
	`
	// Declare a Forum variable to hold the returned data
	var forum Forum
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query using QueryRow()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&forum.ID,
		&forum.User_ID,
		&forum.CreatedAt,
		&forum.Topic,
		&forum.Discussion,
		&forum.Version,
		pq.Array(&forum.Comments),
	)
	// Handle any errors
	if err != nil {
		// Check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Success
	return &forum, nil
}

// Update() allows us to edit/alter a specific Forum
// Optimistic locking (version number)
func (m ForumModel) Update(forum *Forum) error {
	// Create a query
	query := `
		UPDATE forums
		SET topic = $1, discussion = $2, version = version + 1
		WHERE id = $3
		AND version = $4
		RETURNING version
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	args := []interface{}{
		forum.Topic,
		forum.Discussion,
		forum.ID,
		forum.Version,
	}
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&forum.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete() removes a specific Forum
func (m ForumModel) Delete(id int64) error {
	// Ensure that there is a valid id
	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM forums
		WHERE id = $1
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Check how many rows were affected by the delete operation. We
	// call the RowsAffected() method on the result variable
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// Check if no rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// the GetAll() method returns a list of all the schools sorted by id
func (m ForumModel) GetAll(topic string, filters Filters) ([]*Forum, Metadata, error) {
	// Construct the query
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, user_id, created_at, topic, discussion, version, comments
		FROM forums
		WHERE (to_tsvector('simple', topic) @@ plainto_tsquery('simple', $1) OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortOrder())

	// Create a 3-seconds-timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Execute the query
	args := []interface{}{topic, filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	// Close the resultset
	defer rows.Close()
	totalRecords := 0
	// Initialize an empty slice to hold the Forum data
	forums := []*Forum{}
	// Iterate over the rows in the resultset
	for rows.Next() {
		var forum Forum
		// Scan the values from the row into the forum
		err := rows.Scan(
			&totalRecords,
			&forum.ID,
			&forum.User_ID,
			&forum.CreatedAt,
			&forum.Topic,
			&forum.Discussion,
			&forum.Version,
			pq.Array(&forum.Comments),
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		// Add the Forum to our slice
		forums = append(forums, &forum)
	}
	// Check for errors after looping through the resultset
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Return the slice of Forums
	return forums, metadata, nil
}

/*
// addDiscussion() allows us to add a discussion to a specific Forum
// Optimistic locking (version number)
func (m ForumModel) addDiscussion(forum *Forum) error {
	// Create a query
	query := `
		UPDATE forums
		SET discussion = $1, version = version + 1
		WHERE id = $2
		AND version = $3
		RETURNING version
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	args := []interface{}{
		pq.Array(forum.Discussion),
		forum.ID,
		forum.Version,
	}
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&forum.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
} */
