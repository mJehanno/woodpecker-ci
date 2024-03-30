package pipeline

import (
	"context"
	"woodpecker-ci/db"

	"github.com/doug-martin/goqu/v9"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types"
)

type File struct {
	Path string `json:"path" db:"path"`
	Name string `json:"name" db:"filename"`
}

type Pipeline struct {
	File
	ID     int    `json:"id" db:"id"`
	Status string `json:"status" db:"status"`
}

func SavePipeline(ctx context.Context, p Pipeline) error {
	p.Status = "none"
	m := db.Conn.Insert("pipeline").Rows(
		p,
	).Executor()

	_, err := m.ExecContext(ctx)

	return err
}

func ParsePipeline(content string) (*types.Workflow, error) {
	return yaml.ParseString(content)
}

func UpdatePipeline() {}

func DeletePipeline() {}

func GetPipeline(ctx context.Context) ([]Pipeline, error) {
	res := []Pipeline{}
	query := db.Conn.Select(goqu.Star()).From("pipeline")
	err := query.ScanStructsContext(ctx, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}
