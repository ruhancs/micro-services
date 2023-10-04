package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func New(mongo *mongo.Client) Models {
	client = mongo

	return Models {
		LogEntry: LogEntry{},
	}
}

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID string `bson:"_id,omitempty" json:"id,omitempty"`
	Name string `bson:"name" json:"name"`
	Data string `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

//inserir documento no DB
func (l *LogEntry) Insert(entry LogEntry) error {
	//nome da colecao é logs
	collection := client.Database("logs").Collection("logs")

	_,err := collection.InsertOne(context.TODO(), LogEntry{
		Name: entry.Name,
		Data: entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		log.Println("Error inserting into logs: ",err)
		return err
	}

	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	//caso demore mais de 15 segusdos fecha a conexao
	ctx,cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")
	//retornar todos itens da tabela, ordenado por ordem de criacao
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor,err := collection.Find(context.TODO(),bson.D{}, opts)
	if err != nil {
		log.Println("find all doc err: ", err)
		return nil,err
	}

	defer cursor.Close(ctx)

	//variavel para guardar os resultados das buscas
	var logs []*LogEntry

	//inserir os dados encontrados do database em logs
	for cursor.Next(ctx) {
		var item LogEntry
		err = cursor.Decode(&item)
		if err != nil {
			log.Println("error decoding log into slice: ",err)
			return nil, err
		} else {
			logs = append(logs, &item)
		}
	}

	return logs,nil
}

func (l *LogEntry) GetOne(id string) (*LogEntry,error) {
	//caso demore mais de 15 segusdos fecha a conexao
	ctx,cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	//converter o id para o modelo do mongo
	docID,err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil,err
	}
	
	var entry LogEntry
	//buscar o documento pelo id e inserir em entry
	err = collection.FindOne(ctx, bson.M{"_id": docID}).Decode(&entry)
	if err != nil {
		return nil,err
	}

	return &entry,nil
}

func (l *LogEntry) DropCollection() error {
	//caso demore mais de 15 segusdos fecha a conexao
	ctx,cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	//deletar a colecao de logs
	collection := client.Database("logs").Collection("logs")
	if err := collection.Drop(ctx); err != nil {
		return err
	}

	return nil
}

func (l *LogEntry) Update() (*mongo.UpdateResult,error) {
	//caso demore mais de 15 segusdos fecha a conexao
	ctx,cancel := context.WithTimeout(context.Background(), 15 * time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	//converter o id para o modelo do mongo
	docID,err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		return nil,err
	}

	result,err := collection.UpdateOne(
		ctx,
		bson.M{"_id": docID},
		bson.D{
			{"$set", bson.D{
				{"name", l.Name},
				{"data", l.Data},
				{"upadted_at", time.Now()},
			}},
		},
	)
	if err != nil {
		return nil,err
	}
	return result,nil
}