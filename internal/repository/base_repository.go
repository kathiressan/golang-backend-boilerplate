package repository

import "gorm.io/gorm"

// BaseRepository provides generic CRUD operations for any entity type
// T is the entity type (e.g., entities.User, entities.Session)
type BaseRepository[T any] struct {
	db *gorm.DB
}

// NewBaseRepository creates a new generic repository instance
func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

// getDB returns the provided transaction if available, otherwise the base DB connection
func (r *BaseRepository[T]) getDB(tx ...*gorm.DB) *gorm.DB {
	if len(tx) > 0 && tx[0] != nil {
		return tx[0]
	}
	return r.db
}

// Insert new record
func (r *BaseRepository[T]) Create(entity *T, tx ...*gorm.DB) error {
	return r.getDB(tx...).Create(entity).Error
}

// Get a record by ID
func (r *BaseRepository[T]) FindByID(id string, tx ...*gorm.DB) (*T, error) {
	var entity T
	if err := r.getDB(tx...).Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Get a single record matching the conditions
func (r *BaseRepository[T]) FindOne(conditions map[string]any, tx ...*gorm.DB) (*T, error) {
	var entity T
	query := r.getDB(tx...)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	if err := query.First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Get all records matching the condition
func (r *BaseRepository[T]) FindAll(conditions map[string]any, tx ...*gorm.DB) ([]T, error) {
	var entities []T
	query := r.getDB(tx...)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// Get paginated records
func (r *BaseRepository[T]) List(offset, limit int, conditions map[string]any, tx ...*gorm.DB) ([]T, int64, error) {
	var entities []T
	var total int64

	// Use Session to ensure Count and Find operate on independent clones
	query := r.getDB(tx...).Session(&gorm.Session{})
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}

	// Get total count
	var entity T
	if err := query.Model(&entity).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// Update an existing record
func (r *BaseRepository[T]) Update(entity *T, tx ...*gorm.DB) error {
	return r.getDB(tx...).Save(entity).Error
}

// Update specific fields of a record
func (r *BaseRepository[T]) UpdateFields(id string, fields map[string]any, tx ...*gorm.DB) error {
	var entity T
	return r.getDB(tx...).Model(&entity).Where("id = ?", id).Updates(fields).Error
}

// Delete a record by ID
func (r *BaseRepository[T]) Delete(id string, tx ...*gorm.DB) error {
	var entity T
	return r.getDB(tx...).Delete(&entity, "id = ?", id).Error
}

// Delete records matching conditions
func (r *BaseRepository[T]) DeleteWhere(conditions map[string]any, tx ...*gorm.DB) (int64, error) {
	var entity T
	query := r.getDB(tx...)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	result := query.Delete(&entity)
	return result.RowsAffected, result.Error
}

// Count records matching conditions
func (r *BaseRepository[T]) Count(conditions map[string]any, tx ...*gorm.DB) (int64, error) {
	var entity T
	var count int64
	query := r.getDB(tx...).Model(&entity)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	err := query.Count(&count).Error
	return count, err
}

// Checks if a record exists matching conditions
func (r *BaseRepository[T]) Exists(conditions map[string]any, tx ...*gorm.DB) (bool, error) {
	count, err := r.Count(conditions, tx...)
	return count > 0, err
}

// Returns the underlying database instance for custom queries
func (r *BaseRepository[T]) GetDB(tx ...*gorm.DB) *gorm.DB {
	return r.getDB(tx...)
}
