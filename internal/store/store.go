package store

import (
	"context"
	"database/sql"
	"strings"

	"fmt"
	"strconv"

	"github.com/Kirill-Pinyaev/people-api/internal/models"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// ---------- people

func (s *Store) InsertPerson(ctx context.Context,
	firstName string, middleName *string, lastName string,
	gender *string, nationality *string, age *int,
) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO people (first_name, middle_name, last_name, gender, nationality, age)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id
	`, firstName, middleName, lastName, gender, nationality, age).Scan(&id)
	return id, err
}

func (s *Store) GetPersonWithDetails(ctx context.Context, id int64) (models.Person, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, first_name, middle_name, last_name, gender, nationality, age, created_at, updated_at
		FROM people WHERE id=$1
	`, id)
	var p models.Person
	if err := row.Scan(&p.ID, &p.FirstName, &p.MiddleName, &p.LastName, &p.Gender, &p.Nationality, &p.Age, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return models.Person{}, err
	}

	emRows, err := s.db.QueryContext(ctx, `
		SELECT id, person_id, email, is_primary, created_at
		FROM emails WHERE person_id=$1
		ORDER BY is_primary DESC, id ASC
	`, id)
	if err == nil {
		defer emRows.Close()
		for emRows.Next() {
			var e models.Email
			if err := emRows.Scan(&e.ID, &e.PersonID, &e.Email, &e.IsPrimary, &e.CreatedAt); err == nil {
				p.Emails = append(p.Emails, e)
			}
		}
	}

	_ = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM friendships WHERE user_id=$1 OR friend_id=$1
	`, id).Scan(&p.FriendsCount)

	return p, nil
}

func (s *Store) ListPeople(ctx context.Context) ([]models.Person, error) {
	var rows *sql.Rows
	var err error

	rows, err = s.db.QueryContext(ctx, `
            SELECT p.id, p.first_name, p.middle_name, p.last_name, p.gender, p.nationality, p.age, p.created_at, p.updated_at
            FROM people p
			ORDER BY id ASC
        `)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Person
	var ids []int64
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.FirstName, &p.MiddleName, &p.LastName, &p.Gender, &p.Nationality, &p.Age, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
		ids = append(ids, p.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	emailsByPerson, err := s.emailsByPersonIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Emails = emailsByPerson[out[i].ID]
	}
	return out, nil
}

func (s *Store) ListBySurname(ctx context.Context, lastName string) ([]models.Person, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, first_name, middle_name, last_name, gender, nationality, age, created_at, updated_at
		FROM people
		WHERE LOWER(last_name)=LOWER($1)
		ORDER BY id ASC
	`, lastName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Person
	var ids []int64
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.FirstName, &p.MiddleName, &p.LastName, &p.Gender, &p.Nationality, &p.Age, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
		ids = append(ids, p.ID)
	}
	emailsByPerson, err := s.emailsByPersonIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Emails = emailsByPerson[out[i].ID]
	}
	return out, nil
}

func (s *Store) UpdatePerson(ctx context.Context, id int64, req models.UpdatePersonRequest) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE people SET
			first_name = COALESCE($1, first_name),
			middle_name = COALESCE($2, middle_name),
			last_name = COALESCE($3, last_name),
			gender = COALESCE($4, gender),
			nationality = COALESCE($5, nationality),
			age = COALESCE($6, age),
			updated_at = NOW()
		WHERE id=$7
	`, req.FirstName, req.MiddleName, req.LastName, req.Gender, req.Nationality, req.Age, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ---------- emails

func (s *Store) InsertEmail(ctx context.Context, personID int64, email string, isPrimary bool) (int64, error) {
	var id int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO emails (person_id, email, is_primary) VALUES ($1,$2,$3) RETURNING id
	`, personID, email, isPrimary).Scan(&id)
	return id, err
}

func (s *Store) GetEmailByID(ctx context.Context, emailID int64) (models.Email, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, person_id, email, is_primary, created_at FROM emails WHERE id=$1
	`, emailID)
	var e models.Email
	err := row.Scan(&e.ID, &e.PersonID, &e.Email, &e.IsPrimary, &e.CreatedAt)
	return e, err
}

func (s *Store) ListEmails(ctx context.Context, personID int64) ([]models.Email, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, person_id, email, is_primary, created_at
		FROM emails WHERE person_id=$1
		ORDER BY is_primary DESC, id ASC
	`, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Email
	for rows.Next() {
		var e models.Email
		if err := rows.Scan(&e.ID, &e.PersonID, &e.Email, &e.IsPrimary, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

func (s *Store) DeleteEmail(ctx context.Context, personID, emailID int64) (int64, error) {
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM emails WHERE id=$1 AND person_id=$2
	`, emailID, personID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *Store) emailsByPersonIDs(ctx context.Context, ids []int64) (map[int64][]models.Email, error) {
	out := make(map[int64][]models.Email, len(ids))
	if len(ids) == 0 {
		return out, nil
	}

	placeholders := make([]string, 0, len(ids))
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		placeholders = append(placeholders, "$"+strconv.Itoa(i+1))
		args = append(args, id)
	}

	q := fmt.Sprintf(`
        SELECT id, person_id, email, is_primary, created_at
        FROM emails
        WHERE person_id IN (%s)
        ORDER BY is_primary DESC, id ASC
    `, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e models.Email
		if err := rows.Scan(&e.ID, &e.PersonID, &e.Email, &e.IsPrimary, &e.CreatedAt); err != nil {
			return nil, err
		}
		out[e.PersonID] = append(out[e.PersonID], e)
	}
	return out, rows.Err()
}

// ---------- friends

func (s *Store) AddFriend(ctx context.Context, a, b int64) error {
	u1, u2 := a, b
	if u1 > u2 {
		u1, u2 = u2, u1
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO friendships (user_id, friend_id) VALUES ($1,$2) ON CONFLICT DO NOTHING
	`, u1, u2)
	return err
}

func (s *Store) RemoveFriend(ctx context.Context, a, b int64) (int64, error) {
	u1, u2 := a, b
	if u1 > u2 {
		u1, u2 = u2, u1
	}
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM friendships WHERE user_id=$1 AND friend_id=$2
	`, u1, u2)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s *Store) ListFriends(ctx context.Context, id int64) ([]models.Person, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT p.id, p.first_name, p.middle_name, p.last_name, p.gender, p.nationality, p.age, p.created_at, p.updated_at
		FROM people p
		JOIN (
			SELECT CASE WHEN user_id=$1 THEN friend_id ELSE user_id END AS fid
			FROM friendships WHERE user_id=$1 OR friend_id=$1
		) f ON f.fid = p.id
		ORDER BY p.id ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Person
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.FirstName, &p.MiddleName, &p.LastName, &p.Gender, &p.Nationality, &p.Age, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}
