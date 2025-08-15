package demographics

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Service struct {
	client *http.Client
	logger *log.Logger
}

func NewService(c *http.Client) *Service {
	return &Service{client: c}
}

func (s *Service) Infer(ctx context.Context, firstName string) (*int, *string, *string) {
	type ageResp struct {
		Age *int `json:"age"`
	}
	type genderResp struct {
		Gender *string `json:"gender"`
	}
	type natResp struct {
		Country []struct {
			CountryID   string  `json:"country_id"`
			Probability float64 `json:"probability"`
		} `json:"country"`
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	type result[T any] struct {
		v   T
		err error
	}
	ageCh := make(chan result[*int], 1)
	genCh := make(chan result[*string], 1)
	natCh := make(chan result[*string], 1)

	go func() {
		start := time.Now()
		var out *int
		var status int
		url := fmt.Sprintf("https://api.agify.io/?name=%s", firstName)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := s.client.Do(req)
		if resp != nil {
			status = resp.StatusCode
		}
		if err == nil && resp != nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			var ar ageResp
			if json.NewDecoder(resp.Body).Decode(&ar) == nil {
				out = ar.Age
			}
		}
		s.logger.Printf("demographics provider=agify name=%q status=%d dur=%s err=%v",
			firstName, status, time.Since(start), err)
		ageCh <- result[*int]{v: out, err: err}
	}()

	go func() {
		start := time.Now()
		var status int
		var out *string
		url := fmt.Sprintf("https://api.genderize.io/?name=%s", firstName)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := s.client.Do(req)
		if resp != nil {
			status = resp.StatusCode
		}
		if err == nil && resp != nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			var gr genderResp
			if json.NewDecoder(resp.Body).Decode(&gr) == nil {
				out = gr.Gender
			}
		}
		s.logger.Printf("demographics provider=genderize name=%q status=%d dur=%s err=%v",
			firstName, status, time.Since(start), err)
		genCh <- result[*string]{v: out, err: err}
	}()

	go func() {
		start := time.Now()
		var status int
		var out *string
		url := fmt.Sprintf("https://api.nationalize.io/?name=%s", firstName)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := s.client.Do(req)
		if resp != nil {
			status = resp.StatusCode
		}
		if err == nil && resp != nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			var nr natResp
			if json.NewDecoder(resp.Body).Decode(&nr) == nil && len(nr.Country) > 0 {
				best := nr.Country[0]
				for _, c := range nr.Country[1:] {
					if c.Probability > best.Probability {
						best = c
					}
				}
				out = &best.CountryID
			}
		}
		s.logger.Printf("demographics provider=nationalize name=%q status=%d dur=%s err=%v",
			firstName, status, time.Since(start), err)
		natCh <- result[*string]{v: out, err: err}
	}()

	age := (<-ageCh).v
	gender := (<-genCh).v
	nationality := (<-natCh).v
	return age, gender, nationality
}
