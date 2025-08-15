package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Kirill-Pinyaev/people-api/internal/app"
	"github.com/Kirill-Pinyaev/people-api/internal/http/httputil"
	"github.com/Kirill-Pinyaev/people-api/internal/models"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	a *app.App
}

func New(a *app.App) *Handlers {
	return &Handlers{a: a}
}

// --------- People

func (h *Handlers) PeopleCreate(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid json: %v", err)
		return
	}
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	if req.FirstName == "" || req.LastName == "" {
		httputil.Error(w, http.StatusBadRequest, "first_name and last_name are required")
		return
	}

	age, gender, nationality := req.Age, req.Gender, req.Nationality
	if age == nil || gender == nil || nationality == nil {
		a2, g2, n2 := h.a.Demographics.Infer(r.Context(), req.FirstName)
		if age == nil {
			age = a2
		}
		if gender == nil {
			gender = g2
		}
		if nationality == nil {
			nationality = n2
		}
	}

	id, err := h.a.Store.InsertPerson(r.Context(),
		req.FirstName, req.MiddleName, req.LastName, gender, nationality, age,
	)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "insert person: %v", err)
		return
	}

	for _, em := range req.Emails {
		if strings.TrimSpace(em.Email) == "" {
			continue
		}
		if _, err := h.a.Store.InsertEmail(r.Context(), id, em.Email, em.IsPrimary); err != nil {
			httputil.Error(w, http.StatusBadRequest, "insert email %s: %v", em.Email, err)
			return
		}
	}

	person, err := h.a.Store.GetPersonWithDetails(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "get person: %v", err)
		return
	}
	httputil.JSON(w, http.StatusCreated, person)
}

func (h *Handlers) PeopleGet(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id: %v", err)
		return
	}
	p, err := h.a.Store.GetPersonWithDetails(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httputil.Error(w, http.StatusNotFound, "not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "get: %v", err)
		return
	}
	httputil.JSON(w, http.StatusOK, p)
}

func (h *Handlers) PeopleList(w http.ResponseWriter, r *http.Request) {
	out, err := h.a.Store.ListPeople(r.Context())
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "list: %v", err)
		return
	}
	httputil.JSON(w, http.StatusOK, out)
}

func (h *Handlers) PeopleBySurname(w http.ResponseWriter, r *http.Request) {
	lastName := strings.TrimSpace(chi.URLParam(r, "last_name"))
	if lastName == "" {
		httputil.Error(w, http.StatusBadRequest, "last_name required")
		return
	}
	out, err := h.a.Store.ListBySurname(r.Context(), lastName)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "query: %v", err)
		return
	}
	if len(out) == 0 {
		httputil.Error(w, http.StatusNotFound, "no people with surname %s", lastName)
		return
	}
	httputil.JSON(w, http.StatusOK, out)
}

func (h *Handlers) PeopleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id: %v", err)
		return
	}
	var req models.UpdatePersonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid json: %v", err)
		return
	}

	aff, err := h.a.Store.UpdatePerson(r.Context(), id, req)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "update: %v", err)
		return
	}
	if aff == 0 {
		httputil.Error(w, http.StatusNotFound, "not found")
		return
	}
	p, err := h.a.Store.GetPersonWithDetails(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "get: %v", err)
		return
	}
	httputil.JSON(w, http.StatusOK, p)
}

// --------- Emails

func (h *Handlers) AddEmail(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id: %v", err)
		return
	}
	var req struct {
		Email     string `json:"email"`
		IsPrimary bool   `json:"is_primary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid json: %v", err)
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		httputil.Error(w, http.StatusBadRequest, "email required")
		return
	}
	emailID, err := h.a.Store.InsertEmail(r.Context(), id, req.Email, req.IsPrimary)
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "insert: %v", err)
		return
	}
	e, err := h.a.Store.GetEmailByID(r.Context(), emailID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "scan: %v", err)
		return
	}
	httputil.JSON(w, http.StatusCreated, e)
}

func (h *Handlers) ListEmails(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id: %v", err)
		return
	}
	out, err := h.a.Store.ListEmails(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "list: %v", err)
		return
	}
	httputil.JSON(w, http.StatusOK, out)
}

func (h *Handlers) DeleteEmail(w http.ResponseWriter, r *http.Request) {
	personID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id: %v", err)
		return
	}
	emailID, err := parseID(chi.URLParam(r, "email_id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid email_id: %v", err)
		return
	}
	aff, err := h.a.Store.DeleteEmail(r.Context(), personID, emailID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "delete: %v", err)
		return
	}
	if aff == 0 {
		httputil.Error(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --------- Friends

func (h *Handlers) AddFriend(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id: %v", err)
		return
	}
	friendID, err := parseID(chi.URLParam(r, "friend_id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid friend_id: %v", err)
		return
	}
	if id == friendID {
		httputil.Error(w, http.StatusBadRequest, "cannot befriend self")
		return
	}
	if err := h.a.Store.AddFriend(r.Context(), id, friendID); err != nil {
		httputil.Error(w, http.StatusInternalServerError, "insert: %v", err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id: %v", err)
		return
	}
	friendID, err := parseID(chi.URLParam(r, "friend_id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid friend_id: %v", err)
		return
	}
	aff, err := h.a.Store.RemoveFriend(r.Context(), id, friendID)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "delete: %v", err)
		return
	}
	if aff == 0 {
		httputil.Error(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) ListFriends(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		httputil.Error(w, http.StatusBadRequest, "invalid id: %v", err)
		return
	}
	out, err := h.a.Store.ListFriends(r.Context(), id)
	if err != nil {
		httputil.Error(w, http.StatusInternalServerError, "list friends: %v", err)
		return
	}
	httputil.JSON(w, http.StatusOK, out)
}

func parseID(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty id")
	}

	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.New("error parse id")
	}

	if n <= 0 {
		return 0, errors.New("id must be positive")
	}
	return n, nil
}
