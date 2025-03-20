package todo_handler

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Todo struct {
	Id   primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Body string             `json:"body" bson:"body"`
	Done bool               `json:"done" bson:"done"`
}

var collection *mongo.Collection

func GetTodo(c *fiber.Ctx) error {
	var todos []Todo

	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return c.Status(500).JSON(err)
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return c.Status(500).JSON(err)
		}
		todos = append(todos, todo)
	}

	return c.JSON(todos)
}

func GetTodoById(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var todo Todo
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(200).JSON(todo)
}

func CreateTodo(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return c.Status(400).JSON(err)
	}

	inserResult, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		return c.Status(500).JSON(err)
	}

	todo.Id = inserResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func UpdateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	var req Todo

	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{
		"body": req.Body,
		"done": req.Done,
	}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{"success": true})

}

func DeleteTodo(c *fiber.Ctx) error {
	// Parse and validate the ID parameter
	objectID, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid todo ID",
		})
	}

	// Attempt to delete the document
	filter := bson.M{"_id": objectID}
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete todo",
		})
	}

	// Check if a document was actually deleted
	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Todo not found",
		})
	}

	// Respond with success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}
