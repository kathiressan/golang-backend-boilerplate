package repository

import (
	"context"
	"fmt"
	"ovmsa-be/pkg/logger"
	"time"

	"gorm.io/gorm"
)

// RepositoryDecorator defines the interface for repository decorators
// This allows us to wrap repositories with additional functionality
type RepositoryDecorator[T any] interface {
	Create(ctx context.Context, entity *T, tx ...*gorm.DB) error
	FindByID(ctx context.Context, id string, preloads []string, tx ...*gorm.DB) (*T, error)
	FindOne(ctx context.Context, conditions map[string]any, preloads []string, tx ...*gorm.DB) (*T, error)
	FindAll(ctx context.Context, conditions map[string]any, preloads []string, tx ...*gorm.DB) ([]T, error)
	List(ctx context.Context, offset, limit int, sort string, conditions map[string]any, preloads []string, tx ...*gorm.DB) ([]T, int64, error)
	Update(ctx context.Context, entity *T, tx ...*gorm.DB) error
	UpdateFields(ctx context.Context, id string, fields map[string]any, tx ...*gorm.DB) error
	Delete(ctx context.Context, id string, tx ...*gorm.DB) (int64, error)
	DeleteWhere(ctx context.Context, conditions map[string]any, tx ...*gorm.DB) (int64, error)
	Count(ctx context.Context, conditions map[string]any, tx ...*gorm.DB) (int64, error)
	Exists(ctx context.Context, conditions map[string]any, tx ...*gorm.DB) (bool, error)
	ExistsByID(ctx context.Context, id string, tx ...*gorm.DB) (bool, error)
}

// LoggingRepositoryDecorator adds logging to repository operations
type LoggingRepositoryDecorator[T any] struct {
	base       *BaseRepository[T]
	entityName string
}

// NewLoggingRepositoryDecorator creates a new logging decorator
func NewLoggingRepositoryDecorator[T any](base *BaseRepository[T], entityName string) *LoggingRepositoryDecorator[T] {
	return &LoggingRepositoryDecorator[T]{
		base:       base,
		entityName: entityName,
	}
}

// Create logs the creation operation
func (d *LoggingRepositoryDecorator[T]) Create(ctx context.Context, entity *T, tx ...*gorm.DB) error {
	logger.Debug(fmt.Sprintf("Creating %s", d.entityName))
	start := time.Now()
	err := d.base.Create(ctx, entity, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create %s", d.entityName), "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully created %s", d.entityName), "duration", duration)
	}
	return err
}

// FindByID logs the find by ID operation
func (d *LoggingRepositoryDecorator[T]) FindByID(ctx context.Context, id string, preloads []string, tx ...*gorm.DB) (*T, error) {
	logger.Debug(fmt.Sprintf("Finding %s by ID", d.entityName), "id", id)
	start := time.Now()
	result, err := d.base.FindByID(ctx, id, preloads, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to find %s by ID", d.entityName), "id", id, "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully found %s by ID", d.entityName), "id", id, "duration", duration)
	}
	return result, err
}

// FindOne logs the find one operation
func (d *LoggingRepositoryDecorator[T]) FindOne(ctx context.Context, conditions map[string]any, preloads []string, tx ...*gorm.DB) (*T, error) {
	logger.Debug(fmt.Sprintf("Finding one %s", d.entityName), "conditions", conditions)
	start := time.Now()
	result, err := d.base.FindOne(ctx, conditions, preloads, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Debug(fmt.Sprintf("Failed to find %s", d.entityName), "conditions", conditions, "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully found %s", d.entityName), "duration", duration)
	}
	return result, err
}

// FindAll logs the find all operation
func (d *LoggingRepositoryDecorator[T]) FindAll(ctx context.Context, conditions map[string]any, preloads []string, tx ...*gorm.DB) ([]T, error) {
	logger.Debug(fmt.Sprintf("Finding all %s", d.entityName), "conditions", conditions)
	start := time.Now()
	results, err := d.base.FindAll(ctx, conditions, preloads, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to find all %s", d.entityName), "conditions", conditions, "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully found %d %s", len(results), d.entityName), "duration", duration)
	}
	return results, err
}

// List logs the list operation
func (d *LoggingRepositoryDecorator[T]) List(ctx context.Context, offset, limit int, sort string, conditions map[string]any, preloads []string, tx ...*gorm.DB) ([]T, int64, error) {
	logger.Debug(fmt.Sprintf("Listing %s", d.entityName), "offset", offset, "limit", limit, "sort", sort)
	start := time.Now()
	results, total, err := d.base.List(ctx, offset, limit, sort, conditions, preloads, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to list %s", d.entityName), "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully listed %d/%d %s", len(results), total, d.entityName), "duration", duration)
	}
	return results, total, err
}

// Update logs the update operation
func (d *LoggingRepositoryDecorator[T]) Update(ctx context.Context, entity *T, tx ...*gorm.DB) error {
	logger.Debug(fmt.Sprintf("Updating %s", d.entityName))
	start := time.Now()
	err := d.base.Update(ctx, entity, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update %s", d.entityName), "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully updated %s", d.entityName), "duration", duration)
	}
	return err
}

// UpdateFields logs the update fields operation
func (d *LoggingRepositoryDecorator[T]) UpdateFields(ctx context.Context, id string, fields map[string]any, tx ...*gorm.DB) error {
	logger.Debug(fmt.Sprintf("Updating %s fields", d.entityName), "id", id, "fields", fields)
	start := time.Now()
	err := d.base.UpdateFields(ctx, id, fields, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update %s fields", d.entityName), "id", id, "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully updated %s fields", d.entityName), "id", id, "duration", duration)
	}
	return err
}

// Delete logs the delete operation
func (d *LoggingRepositoryDecorator[T]) Delete(ctx context.Context, id string, tx ...*gorm.DB) (int64, error) {
	logger.Debug(fmt.Sprintf("Deleting %s", d.entityName), "id", id)
	start := time.Now()
	rows, err := d.base.Delete(ctx, id, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to delete %s", d.entityName), "id", id, "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully deleted %s", d.entityName), "id", id, "rows_affected", rows, "duration", duration)
	}
	return rows, err
}

// DeleteWhere logs the delete where operation
func (d *LoggingRepositoryDecorator[T]) DeleteWhere(ctx context.Context, conditions map[string]any, tx ...*gorm.DB) (int64, error) {
	logger.Debug(fmt.Sprintf("Deleting %s where", d.entityName), "conditions", conditions)
	start := time.Now()
	rows, err := d.base.DeleteWhere(ctx, conditions, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to delete %s where", d.entityName), "conditions", conditions, "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully deleted %s where", d.entityName), "rows_affected", rows, "duration", duration)
	}
	return rows, err
}

// Count logs the count operation
func (d *LoggingRepositoryDecorator[T]) Count(ctx context.Context, conditions map[string]any, tx ...*gorm.DB) (int64, error) {
	logger.Debug(fmt.Sprintf("Counting %s", d.entityName), "conditions", conditions)
	start := time.Now()
	count, err := d.base.Count(ctx, conditions, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to count %s", d.entityName), "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Successfully counted %s", d.entityName), "count", count, "duration", duration)
	}
	return count, err
}

// Exists logs the exists operation
func (d *LoggingRepositoryDecorator[T]) Exists(ctx context.Context, conditions map[string]any, tx ...*gorm.DB) (bool, error) {
	logger.Debug(fmt.Sprintf("Checking if %s exists", d.entityName), "conditions", conditions)
	start := time.Now()
	exists, err := d.base.Exists(ctx, conditions, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to check if %s exists", d.entityName), "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Checked if %s exists", d.entityName), "exists", exists, "duration", duration)
	}
	return exists, err
}

// ExistsByID logs the exists by ID operation
func (d *LoggingRepositoryDecorator[T]) ExistsByID(ctx context.Context, id string, tx ...*gorm.DB) (bool, error) {
	logger.Debug(fmt.Sprintf("Checking if %s exists by ID", d.entityName), "id", id)
	start := time.Now()
	exists, err := d.base.ExistsByID(ctx, id, tx...)
	duration := time.Since(start)
	
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to check if %s exists by ID", d.entityName), "id", id, "error", err, "duration", duration)
	} else {
		logger.Debug(fmt.Sprintf("Checked if %s exists by ID", d.entityName), "id", id, "exists", exists, "duration", duration)
	}
	return exists, err
}

// WithLogging wraps a repository with logging decorator
func WithLogging[T any](repo *BaseRepository[T], entityName string) *LoggingRepositoryDecorator[T] {
	return NewLoggingRepositoryDecorator(repo, entityName)
}
