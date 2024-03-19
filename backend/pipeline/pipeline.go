package pipeline

import (
	"context"
	"woodpecker-ci/db"

	"github.com/doug-martin/goqu/v9"
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

func SavePipeline() {
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
