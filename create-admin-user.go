package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
	Role      string             `bson:"role"`
}

func main() {
	log.Println("Conectando ao MongoDB...")
	
	// Conectar ao MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27185")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Erro ao conectar ao MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	// Verificar a conexão
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Erro ao verificar conexão com MongoDB: %v", err)
	}
	log.Println("Conectado ao MongoDB com sucesso!")

	// Acessar a coleção
	collection := client.Database("identity").Collection("users")

	// Verificar se o usuário já existe
	var existingUser User
	err = collection.FindOne(context.Background(), bson.M{"email": "admin@example.com"}).Decode(&existingUser)
	
	if err == nil {
		// Usuário existe, atualizar a senha
		log.Println("Usuário admin@example.com já existe, atualizando senha...")
		
		// Hash da senha
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Erro ao gerar hash da senha: %v", err)
		}
		
		// Atualizar usuário
		filter := bson.M{"email": "admin@example.com"}
		update := bson.M{
			"$set": bson.M{
				"password": string(hashedPassword),
				"updatedAt": time.Now(),
			},
		}
		
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			log.Fatalf("Erro ao atualizar senha: %v", err)
		}
		
		log.Println("Senha atualizada com sucesso!")
	} else if err == mongo.ErrNoDocuments {
		// Usuário não existe, criar novo
		log.Println("Usuário admin@example.com não existe, criando...")
		
		// Hash da senha
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Erro ao gerar hash da senha: %v", err)
		}
		
		// Criar novo usuário
		newUser := User{
			Name:      "Admin",
			Email:     "admin@example.com",
			Password:  string(hashedPassword),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Role:      "admin", // Define como admin
		}
		
		result, err := collection.InsertOne(context.Background(), newUser)
		if err != nil {
			log.Fatalf("Erro ao inserir usuário: %v", err)
		}
		
		log.Printf("Usuário criado com sucesso! ID: %v", result.InsertedID)
	} else {
		log.Fatalf("Erro ao verificar usuário: %v", err)
	}
	
	fmt.Println("\nUsuário admin@example.com configurado com sucesso!")
	fmt.Println("Email: admin@example.com")
	fmt.Println("Senha: password123")
}
