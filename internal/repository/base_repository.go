package repository

import (
	"context"
	"errors"
	"fmt"
	appErrors "ovmsa-be/pkg/errors"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// BaseRepository provides generic CRUD operations for any entity type
// T is the entity type (e.g., entities.User, entities.Session)
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new generic repository instance
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// wrapError is a helper to convert GORM errors into AppErrors
func (r *BaseRepository[T]) wrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.NotFound(err, message)
	}
	return appErrors.InternalServerError(err, message)
}

// sanitizeIdentifier checks if a string is a valid SQL identifier or sort expression
// This prevents SQL injection while allowing common patterns like table aliases and multi-column sorting.
func (r *BaseRepository[T]) sanitizeIdentifier(identifier string) error {
	if identifier == "" {
		return nil
	}
	
	// Split by comma for multi-column sorting
	parts := strings.Split(identifier, ",")
	
	// Standard identifier regex: alphanumeric, underscores, dots, and optional ASC/DESC
	// Case-insensitive to support standard Go field names (e.g., CreatedAt)
	valid := regexp.MustCompile(`(?i)^[a-z0-9_\.]+(\s+(asc|desc))?$`)
	
	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			continue
		}
		if !valid.MatchString(trimmedPart) {
			return fmt.Errorf("invalid identifier: %s", identifier)
		}
	}
	return nil
}

// validateConditions ensures all keys in the conditions map are safe identifiers
func (r *BaseRepository[T]) validateConditions(conditions map[string]any) error {
	for key := range conditions {
		if err := r.sanitizeIdentifier(key); err != nil {
			return err
		}
	}
	return nil
}

// getDB returns the provided transaction if available, otherwise the base DB connection.
// It automatically applies the provided context for RLS and tracing.
func (r *BaseRepository[T]) getDB(ctx context.Context, tx ...*gorm.DB) *gorm.DB {
	base := r.db
	if len(tx) > 0 && tx[0] != nil {
		base = tx[0]
	}
	return base.WithContext(ctx)
}

// applyPreloads is a helper to apply GORM preloads to a query
func (r *BaseRepository[T]) applyPreloads(db *gorm.DB, preloads []string) *gorm.DB {
	for _, preload := range preloads {
		db = db.Preload(preload)
	}
	return db
}

// Insert new record
func (r *BaseRepository[T]) Create(ctx context.Context, entity *T, tx ...*gorm.DB) error {
	err := r.getDB(ctx, tx...).Create(entity).Error
	return r.wrapError(err, "Failed to create record")
}

// Get a record by ID
func (r *BaseRepository[T]) FindByID(ctx context.Context, id string, preloads []string, tx ...*gorm.DB) (*T, error) {
	var entity T
	db := r.getDB(ctx, tx...)
	db = r.applyPreloads(db, preloads)
	
	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, r.wrapError(err, "Record not found")
	}
	return &entity, nil
}

// Get a single record matching the conditions
func (r *BaseRepository[T]) FindOne(ctx context.Context, conditions map[string]any, preloads []string, tx ...*gorm.DB) (*T, error) {
	var entity T
	query := r.getDB(ctx, tx...)
	query = r.applyPreloads(query, preloads)
	
	if err := r.validateConditions(conditions); err != nil {
		return nil, appErrors.BadRequest(err, "Invalid query conditions")
	}
	
	if err := query.Where(conditions).First(&entity).Error; err != nil {
		return nil, r.wrapError(err, "Record not found")
	}
	return &entity, nil
}

// Get all records matching the condition
func (r *BaseRepository[T]) FindAll(ctx context.Context, conditions map[string]any, preloads []string, tx ...*gorm.DB) ([]T, error) {
	var entities []T
	query := r.getDB(ctx, tx...)
	query = r.applyPreloads(query, preloads)
	
	if err := r.validateConditions(conditions); err != nil {
		return nil, appErrors.BadRequest(err, "Invalid query conditions")
	}
	
	if err := query.Where(conditions).Find(&entities).Error; err != nil {
		return nil, r.wrapError(err, "Failed to fetch records")
	}
	return entities, nil
}

// Get paginated records
func (r *BaseRepository[T]) List(ctx context.Context, offset, limit int, sort string, conditions map[string]any, preloads []string, tx ...*gorm.DB) ([]T, int64, error) {
	var entities []T
	var total int64

	// Use Session to ensure Count and Find operate on independent clones
	query := r.getDB(ctx, tx...).Session(&gorm.Session{})
	
	if err := r.validateConditions(conditions); err != nil {
		return nil, 0, appErrors.BadRequest(err, "Invalid query conditions")
	}
	query = query.Where(conditions)

	// Get total count
	var entity T
	if err := query.Model(&entity).Count(&total).Error; err != nil {
		return nil, 0, r.wrapError(err, "Failed to count records")
	}

	// Apply preloads ONLY for the data fetch, not the count
	query = r.applyPreloads(query, preloads)

	// Apply sorting if provided (with sanitization)
	if sort != "" {
		if err := r.sanitizeIdentifier(sort); err != nil {
			return nil, 0, appErrors.BadRequest(err, "Invalid sort parameter")
		}
		query = query.Order(sort)
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, r.wrapError(err, "Failed to fetch paginated records")
	}

	return entities, total, nil
}

// Update an existing record (Full update)
func (r *BaseRepository[T]) Update(ctx context.Context, entity *T, tx ...*gorm.DB) error {
	// We use Updates on the model to ensure GORM handles hooks and only updates valid fields.
	// NOTE: This uses GORM's struct updates, which ignores zero-values (0, "", false).
	// For zero-value updates, use UpdateFields with a map.
	err := r.getDB(ctx, tx...).Model(entity).Updates(entity).Error
	return r.wrapError(err, "Failed to update record")
}

// Update specific fields of a record (Partial update)
func (r *BaseRepository[T]) UpdateFields(ctx context.Context, id string, fields map[string]any, tx ...*gorm.DB) error {
	var entity T
	if err := r.validateConditions(fields); err != nil {
		return appErrors.BadRequest(err, "Invalid field names")
	}
	err := r.getDB(ctx, tx...).Model(&entity).Where("id = ?", id).Updates(fields).Error
	return r.wrapError(err, "Failed to update specific fields")
}

// Delete a record by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id string, tx ...*gorm.DB) (int64, error) {
	var entity T
	result := r.getDB(ctx, tx...).Delete(&entity, "id = ?", id)
	if result.Error != nil {
		return 0, r.wrapError(result.Error, "Failed to delete record")
	}
	return result.RowsAffected, nil
}

// Delete records matching conditions
func (r *BaseRepository[T]) DeleteWhere(ctx context.Context, conditions map[string]any, tx ...*gorm.DB) (int64, error) {
	var entity T
	query := r.getDB(ctx, tx...)
	if err := r.validateConditions(conditions); err != nil {
		return 0, appErrors.BadRequest(err, "Invalid query conditions")
	}
	result := query.Where(conditions).Delete(&entity)
	if result.Error != nil {
		return 0, r.wrapError(result.Error, "Failed to delete records")
	}
	return result.RowsAffected, nil
}

// Count records matching conditions
func (r *BaseRepository[T]) Count(ctx context.Context, conditions map[string]any, tx ...*gorm.DB) (int64, error) {
	var entity T
	var count int64
	query := r.getDB(ctx, tx...).Model(&entity)
	if err := r.validateConditions(conditions); err != nil {
		return 0, appErrors.BadRequest(err, "Invalid query conditions")
	}
	err := query.Where(conditions).Count(&count).Error
	return count, r.wrapError(err, "Failed to count records")
}

// Checks if a record exists matching conditions
func (r *BaseRepository[T]) Exists(ctx context.Context, conditions map[string]any, tx ...*gorm.DB) (bool, error) {
	count, err := r.Count(ctx, conditions, tx...)
	return count > 0, err
}

// ExistsByID checks if a record exists by its primary key
func (r *BaseRepository[T]) ExistsByID(ctx context.Context, id string, tx ...*gorm.DB) (bool, error) {
	return r.Exists(ctx, map[string]any{"id": id}, tx...)
}

// Returns the underlying database instance for custom queries
func (r *BaseRepository[T]) GetDB(ctx context.Context, tx ...*gorm.DB) *gorm.DB {
	return r.getDB(ctx, tx...)
}
