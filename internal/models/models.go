package models

import "time"

// Domain models

type Person struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	MiddleName   *string   `json:"middle_name,omitempty"`
	LastName     string    `json:"last_name"`
	Gender       *string   `json:"gender,omitempty"`
	Nationality  *string   `json:"nationality,omitempty"`
	Age          *int      `json:"age,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Emails       []Email   `json:"emails,omitempty"`
	FriendsCount int       `json:"friends_count,omitempty"`
}

type Email struct {
	ID        int64     `json:"id"`
	PersonID  int64     `json:"person_id"`
	Email     string    `json:"email"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

// Requests

type CreatePersonRequest struct {
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name"`
	LastName   string  `json:"last_name"`

	Gender      *string `json:"gender"`
	Nationality *string `json:"nationality"`
	Age         *int    `json:"age"`

	Emails []struct {
		Email     string `json:"email"`
		IsPrimary bool   `json:"is_primary"`
	} `json:"emails"`
}

type UpdatePersonRequest struct {
	FirstName   *string `json:"first_name"`
	MiddleName  *string `json:"middle_name"`
	LastName    *string `json:"last_name"`
	Gender      *string `json:"gender"`
	Nationality *string `json:"nationality"`
	Age         *int    `json:"age"`
}
