package repositories

import (
	"context"
	"errors"
	"zatrano/pkg/customerrors"

	"gorm.io/gorm"
)

type BaseRepository[T any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any](db *gorm.DB) *BaseRepository[T] {
	return &BaseRepository[T]{db: db}
}

func (r *BaseRepository[T]) GetAll() ([]T, error) {
	var items []T
	err := r.db.Find(&items).Error
	return items, err
}

func (r *BaseRepository[T]) GetByID(id uint) (*T, error) {
	var item T
	err := r.db.First(&item, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, customerrors.ErrRepoRecordNotFound
	}
	return &item, err
}

func (r *BaseRepository[T]) GetCount() (int64, error) {
	var count int64
	err := r.db.Model(new(T)).Count(&count).Error
	return count, err
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *BaseRepository[T]) Update(ctx context.Context, id uint, data map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Updates(data)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return customerrors.ErrRepoRecordNotFound
	}
	return nil
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T

	findTx := r.db.WithContext(ctx).First(&entity, id)
	if findTx.Error != nil {
		if errors.Is(findTx.Error, gorm.ErrRecordNotFound) {
			return customerrors.ErrRepoRecordNotFound
		}
		return findTx.Error
	}

	userID, ok := ctx.Value("user_id").(uint)
	if !ok || userID == 0 {
		return customerrors.ErrInvalidUserContext
	}

	if r.db.Migrator().HasColumn(&entity, "deleted_by") {
		updateTx := r.db.WithContext(ctx).Model(&entity).Update("deleted_by", userID)
		if updateTx.Error != nil {
			return updateTx.Error
		}
	}

	deleteTx := r.db.WithContext(ctx).Delete(&entity)
	if deleteTx.Error != nil {
		return deleteTx.Error
	}
	if deleteTx.RowsAffected == 0 {
		return customerrors.ErrRepoRecordNotFound
	}

	return nil
}
