package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User representa a estrutura de um usuário no sistema
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name" json:"name" binding:"required"`
	Email    string             `bson:"email" json:"email" binding:"required,email"`
	Password string             `bson:"password" json:"password" binding:"required,min=8"`
	// Adicionar outros campos conforme necessário, como IsActive, Roles, Timestamps etc.
}
