package main

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"
)

func NewVectorStore(db *sqlx.DB) *VectorStore {
	return &VectorStore{
		db:      db,
		dbCache: sq.NewStmtCache(db),
	}
}

type VectorStore struct {
	db      *sqlx.DB
	dbCache *sq.StmtCache
}

func (s *VectorStore) QueryContent(ctx context.Context, query []float32, similarityThreshold float32) ([]string, error) {
	sb := sq.StatementBuilder.RunWith(s.dbCache)
	selectCourses := sb.
		// Select("langchain_id", "content", "embedding", "langchain_metadata").
		Select("content", "1 - (c.embedding <=> $1) as distance").
		From("course_content_embeddings c").
		Where("1 - (c.embedding <=> $1) > $2",
			pgvector.NewVector(query), similarityThreshold).
		OrderBy("distance desc").
		Limit(5).
		PlaceholderFormat(sq.Dollar)

	rows, err := selectCourses.QueryContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("query content error")
		return nil, err
	}

	var contents []string
	for rows.Next() {
		var content string
		var distance float32
		if err := rows.Scan(&content, &distance); err != nil {
			log.Error().Err(err).Msg("scan content error")
			return nil, err
		}
		// log.Debug().Str("content", content).Float32("distance", distance).Msg("content")
		contents = append(contents, content)
	}
	return contents, nil
}
