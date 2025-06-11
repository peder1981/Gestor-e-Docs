package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User representa a estrutura de um usuário no sistema
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name" binding:"required"`
	Email     string             `bson:"email" json:"email" binding:"required,email"`
	Password  string             `bson:"password" json:"-" binding:"required,min=8"`
	Role      string             `bson:"role,omitempty" json:"role,omitempty"`
	CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt,omitempty"`
}
