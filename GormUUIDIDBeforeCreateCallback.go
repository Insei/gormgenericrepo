package gormgenericrepo

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
)

func reflectSetUUIDForSchemeField(fieldScheme *schema.Field, ctx context.Context, value reflect.Value) {
	idAlreadySet := false
	if fieldValue, isZero := fieldScheme.ValueOf(ctx, value); !isZero {
		if id, ok := fieldValue.(uuid.UUID); ok {
			if id != uuid.Nil {
				idAlreadySet = true
			}
		}
	}
	if !idAlreadySet {
		err := fieldScheme.Set(ctx, value, uuid.New())
		if err != nil {
			panic(err)
		}
	}
}

// beforeCreate hook with uuid.UUID id set
func beforeCreate(db *gorm.DB) {
	if db.Statement.Schema != nil && db.Statement.Schema.PrioritizedPrimaryField != nil {
		idField := db.Statement.Schema.PrioritizedPrimaryField
		switch idField.FieldType {
		case reflect.TypeOf(uuid.UUID{}):
			switch db.Statement.ReflectValue.Kind() {
			case reflect.Slice, reflect.Array:
				for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
					refVal := db.Statement.ReflectValue.Index(i)
					reflectSetUUIDForSchemeField(idField, db.Statement.Context, refVal)
				}
			default:
				reflectSetUUIDForSchemeField(idField, db.Statement.Context, db.Statement.ReflectValue)
			}
		}
	}
}

// AddGormBeforeCreateUUIDIDSetCallback sets before "gorm:before_create" callback that sets uuid ids for entities
func AddGormBeforeCreateUUIDIDSetCallback(db *gorm.DB) {
	err := db.Callback().Create().Before("gorm:before_create").Replace("uuid:before_create", beforeCreate)
	if err != nil {
		panic(err)
	}
}
