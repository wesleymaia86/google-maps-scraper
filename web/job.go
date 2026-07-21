package web

import (
	"context"
	"errors"
	"time"
)

var jobs []Job

const (
	StatusPending = "pending"
	StatusWorking = "working"
	StatusOK      = "ok"
	StatusFailed  = "failed"
)

type SelectParams struct {
	Status string
	Limit  int
}

type JobRepository interface {
	Get(context.Context, string) (Job, error)
	Create(context.Context, *Job) error
	Delete(context.Context, string) error
	Select(context.Context, SelectParams) ([]Job, error)
	Update(context.Context, *Job) error
}

type Job struct {
	ID     string
	Name   string
	Date   time.Time
	Status string
	Data   JobData
}

func (j *Job) Validate() error {
	if j.ID == "" {
		return errors.New("id ausente")
	}

	if j.Name == "" {
		return errors.New("informe o nome do job")
	}

	if j.Status == "" {
		return errors.New("status ausente")
	}

	if j.Date.IsZero() {
		return errors.New("data ausente")
	}

	if err := j.Data.Validate(); err != nil {
		return err
	}

	return nil
}

type JobData struct {
	Keywords     []string      `json:"keywords"`
	Lang         string        `json:"lang"`
	Zoom         int           `json:"zoom"`
	Lat          string        `json:"lat"`
	Lon          string        `json:"lon"`
	FastMode     bool          `json:"fast_mode"`
	Radius       int           `json:"radius"`
	Depth        int           `json:"depth"`
	Email        bool          `json:"email"`
	ExtraReviews bool          `json:"extra_reviews"`
	MaxTime      time.Duration `json:"max_time"`
	Proxies      []string      `json:"proxies"`
	ExcludeJobID string        `json:"exclude_job_id,omitempty"`
}

func (d *JobData) Validate() error {
	if len(d.Keywords) == 0 {
		return errors.New("informe ao menos uma palavra-chave")
	}

	if d.Lang == "" {
		return errors.New("informe o idioma")
	}

	if len(d.Lang) != 2 {
		return errors.New("idioma deve ter 2 letras (ex.: pt)")
	}

	if d.Depth == 0 {
		return errors.New("informe a profundidade")
	}

	if d.MaxTime == 0 {
		return errors.New("informe o tempo máximo")
	}

	if d.FastMode && (d.Lat == "" || d.Lon == "") {
		return errors.New("modo rápido exige latitude e longitude")
	}

	return nil
}
