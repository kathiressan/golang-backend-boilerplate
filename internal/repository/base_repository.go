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

// Insert new record
func (r *BaseRepository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

// Get a record by ID
func (r *BaseRepository[T]) FindByID(id string) (*T, error) {
	var entity T
	if err := r.db.Where("id = ?", id).First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Get a single record matching the conditions
func (r *BaseRepository[T]) FindOne(conditions map[string]any) (*T, error) {
	var entity T
	query := r.db
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	if err := query.First(&entity).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// Get all records matching the condition
func (r *BaseRepository[T]) FindAll(conditions map[string]any) ([]T, error) {
	var entities []T
	query := r.db
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// Get paginated records
func (r *BaseRepository[T]) List(offset, limit int, conditions map[string]any) ([]T, int64, error) {
	var entities []T
	var total int64

	query := r.db
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
func (r *BaseRepository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

// Update specific fields of a record
func (r *BaseRepository[T]) UpdateFields(id string, fields map[string]any) error {
	var entity T
	return r.db.Model(&entity).Where("id = ?", id).Updates(fields).Error
}

// Delete a record by ID
func (r *BaseRepository[T]) Delete(id string) error {
	var entity T
	return r.db.Delete(&entity, "id = ?", id).Error
}

// Delete records matching conditions
func (r *BaseRepository[T]) DeleteWhere(conditions map[string]any) (int64, error) {
	var entity T
	query := r.db
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	result := query.Delete(&entity)
	return result.RowsAffected, result.Error
}

// Count records matching conditions
func (r *BaseRepository[T]) Count(conditions map[string]any) (int64, error) {
	var entity T
	var count int64
	query := r.db.Model(&entity)
	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}
	err := query.Count(&count).Error
	return count, err
}

// Checks if a record exists matching conditions
func (r *BaseRepository[T]) Exists(conditions map[string]any) (bool, error) {
	count, err := r.Count(conditions)
	return count > 0, err
}

// Returns the underlying database instance for custom queries
func (r *BaseRepository[T]) GetDB() *gorm.DB {
	return r.db
}