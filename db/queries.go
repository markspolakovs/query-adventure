package db

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/couchbase/gocb/v2"
)

func (c *CBDatabase) ExecuteQuery(ctx context.Context, keyspace, query string) ([]any, error) {
	bucket, scope, ok := strings.Cut(keyspace, ".")
	if !ok {
		return nil, fmt.Errorf("invalid keyspace %q", keyspace)
	}
	qr, err := c.cluster.Bucket(bucket).Scope(scope).Query(query, &gocb.QueryOptions{
		Context: ctx,
		Adhoc:   true,
		Timeout: c.queryTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	var rows []any
	for qr.Next() {
		var row any
		err = qr.Row(&row)
		if err != nil {
			return nil, fmt.Errorf("row error: %w", err)
		}
		rows = append(rows, row)
	}
	err = qr.Close()
	if err != nil {
		return nil, fmt.Errorf("close error: %w", err)
	}
	return rows, nil
}

func (c *CBDatabase) ExecuteAndVerifyQuery(ctx context.Context, keyspace, target, input string) error {
	bucket, scope, ok := strings.Cut(keyspace, ".")
	if !ok {
		return fmt.Errorf("invalid keyspace %q", keyspace)
	}
	ks := c.cluster.Bucket(bucket).Scope(scope)
	targetQR, err := ks.Query(target, &gocb.QueryOptions{
		Context: ctx,
		Adhoc:   true,
		Timeout: c.queryTimeout,
	})
	if err != nil {
		return fmt.Errorf("query 1 error: %w", err)
	}
	inputQR, err := ks.Query(input, &gocb.QueryOptions{
		Context: ctx,
		Adhoc:   true,
		Timeout: c.queryTimeout,
	})
	if err != nil {
		return fmt.Errorf("query 2 error: %w", err)
	}

	var targetRows, inputRows uint
	var finalErr error
	var targetRow, inputRow any
	for targetQR.Next() {
		targetRows++
		err = targetQR.Row(&targetRow)
		if err != nil {
			return fmt.Errorf("failed to parse row from target: %w", err)
		}

		ok := inputQR.Next()
		if !ok {
			goto notEnough
		}
		inputRows++
		err = inputQR.Row(&inputRow)
		if err != nil {
			return fmt.Errorf("failed to parse row from input: %w", err)
		}

		if !reflect.DeepEqual(targetRow, inputRow) {
			finalErr = errMismatch(inputRows, targetRow, inputRow)
			goto exit
		}
	}
	if inputQR.Next() {
		inputRows++
		err = inputQR.Row(&inputRow)
		if err != nil {
			return fmt.Errorf("failed to parse row from input (in too many rows loop): %w", err)
		}
		for inputQR.Next() {
			inputRows++
		}
		finalErr = errTooManyRows(targetRows, inputRows, targetRow, inputRow)
		goto exit
	}
	finalErr = nil
	goto exit
notEnough:
	// Run target to the end to get the expected number
	for targetQR.Next() {
		targetRows++
	}
	finalErr = errNotEnoughRows(targetRows, inputRows, inputRow, targetRow)
	goto exit
exit:
	err = targetQR.Close()
	err = inputQR.Close()
	return finalErr
}
