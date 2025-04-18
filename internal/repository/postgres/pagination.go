package postgres

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/vars7899/iots/pkg/apperror"
	"github.com/vars7899/iots/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QueryOptions struct {
	SelectField          []string
	Preloads             map[string][]interface{}
	Timeout              time.Duration
	Hints                []string
	ForUpdate            bool
	SkipCount            bool
	SkipData             bool
	UserMaterializedView bool
}

func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{
		Timeout:   500 * time.Millisecond,
		Preloads:  make(map[string][]interface{}),
		Hints:     make([]string, 0),
		ForUpdate: false,
		SkipCount: false,
	}
}

func FindWithPagination[T any](ctx context.Context, db *gorm.DB, p *pagination.Pagination, queryBuilder func(*gorm.DB) *gorm.DB, log *zap.Logger, opts ...*QueryOptions) ([]T, int64, error) {

	start := time.Now()

	var modelInstance T
	modelType := getModelTypeName(modelInstance)

	var (
		result []T
		count  int64
		err    error
	)

	// process options (if provided)
	var options *QueryOptions
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	} else {
		options = DefaultQueryOptions()
	}

	// apply timeout if valid
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	tx := db.WithContext(ctx) // bind parent/derived context

	// separate db session for count and result
	countTx := queryBuilder(tx.Session(&gorm.Session{NewDB: true}).Model(new(T)))
	dataTx := queryBuilder(tx.Session(&gorm.Session{NewDB: true}).Model(new(T)))

	// apply select field
	if len(options.SelectField) > 0 {
		dataTx = dataTx.Select(options.SelectField)
	}

	// add hints
	for _, hint := range options.Hints {
		countTx = countTx.Clauses(clause.Expr{SQL: hint})
		dataTx = dataTx.Clauses(clause.Expr{SQL: hint})
	}

	// apply for updates
	for options.ForUpdate {
		dataTx = dataTx.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	if !options.SkipCount {
		countQuery := countTx.Order("")
		if options.UserMaterializedView {
			countQuery = countQuery.Table(fmt.Sprintf("%s_materialized", strings.ToLower(modelType)))
		}

		countErrChan := make(chan error, 1)
		go func() {
			countErrChan <- countQuery.Count(&count).Error
		}()

		select {
		case err = <-countErrChan:
			return nil, 0, apperror.MapDBError(err, modelType)
		case <-ctx.Done():
			return nil, 0, apperror.ErrContextTimeout.WithMessage("query timed out before completion due to context timeout").Wrap(err)
		}

	}

	// perform data query
	if !options.SkipData {
		// add preloads
		for relation, args := range options.Preloads {
			if len(args) > 0 {
				dataTx = dataTx.Preload(relation, args...)
			} else {
				dataTx = dataTx.Preload(relation)
			}
		}
		// add pagination & order
		dataTx = dataTx.Offset(p.GetOffset()).Limit(p.GetLimit()).Order(p.GetSortOrderClause())

		dataErrChan := make(chan error, 1)
		go func() {
			dataErrChan <- dataTx.Find(&result).Error
		}()

		select {
		case err = <-dataErrChan:
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Return empty slice instead of error for "not found"
					return []T{}, 0, nil
				}
				return nil, 0, apperror.MapDBError(err, modelType)
			}
		case <-ctx.Done():
			return nil, 0, apperror.ErrContextTimeout.WithMessage("query timed out before completion due to context timeout").Wrap(err)
		}
	}

	// Log slow queries
	duration := time.Since(start)
	if duration > 500*time.Millisecond {
		log.Warn("Slow pagination query detected",
			zap.String("model", modelType),
			zap.Int("limit", p.GetLimit()),
			zap.Int("offset", p.GetOffset()),
			zap.String("sort", p.GetSortOrderClause()),
			zap.Duration("duration", duration),
			zap.Int64("resultCount", int64(len(result))),
			zap.Int64("totalCount", count))
	}

	// update pagination metadata
	p.SetTotalEntries(count)
	if p.PageSize > 0 {
		p.SetPage((count + int64(p.PageSize) - 1) / int64(p.PageSize))
	}

	return result, count, nil
}

func getModelTypeName(modelInstance interface{}) string {
	t := reflect.TypeOf(modelInstance)

	// If it's a pointer, get the underlying element type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Return the base type name
	return t.Name()
}
