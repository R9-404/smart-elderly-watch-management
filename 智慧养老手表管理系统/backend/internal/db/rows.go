package db

import "context"

// QueryList scans all rows and closes the result before returning to the caller.
func QueryList[T any](ctx context.Context, q interface {
	Query(context.Context, string, ...any) (Rows, error)
}, scan func(Rows) (T, error), query string, args ...any) ([]T, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []T{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ScanValue[T any](row Rows) (T, error) {
	var value T
	err := row.Scan(&value)
	return value, err
}
