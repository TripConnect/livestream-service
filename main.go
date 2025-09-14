package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/gocql/gocql"
	"github.com/kristoiv/gocqltable"
	"github.com/tripconnect/go-common-utils/common"
	"github.com/tripconnect/go-common-utils/helper"
	"github.com/tripconnect/go-proto-lib/protos"
	"github.com/tripconnect/livestream-service/consts"
	"github.com/tripconnect/livestream-service/model"
	"github.com/tripconnect/livestream-service/rpc"
	"google.golang.org/grpc"
)

func initCassandra() {
	host, hostErr := helper.ReadConfig[string]("database.cassandra.host")
	username, usernameErr := helper.ReadConfig[string]("database.cassandra.username")
	password, passwordErr := helper.ReadConfig[string]("database.cassandra.password")

	if hostErr != nil || usernameErr != nil || passwordErr != nil {
		log.Fatal("failed to load cassandra config")
	}

	// Authentication
	cluster := gocql.NewCluster(host)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra: %v", err)
	}
	gocqltable.SetDefaultSession(session)

	// Create keyspace
	keyspace := gocqltable.NewKeyspace(consts.KeySpace)
	_ = keyspace.Create(map[string]interface{}{
		"class":              "SimpleStrategy",
		"replication_factor": 1,
	}, true)

	// Create tables
	model.LivestreamRepository.TableInterface.Create()

}

func initElasticsearch() {
	ctx := context.Background()
	// Create indexes

	common.ElasticsearchClient.Indices.
		Create(consts.LivestreamIndex).
		Mappings(model.LivestreamDocumentMappings).
		Do(ctx)
}

func init() {
	// Cassandra initalization
	initCassandra()
	// Elasticsearch initalization
	initElasticsearch()
	// Kafka initalization
}

func main() {
	port, err := helper.ReadConfig[int]("server.port")

	if err != nil {
		log.Fatalf("failed to load port config %v", err)
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var server = grpc.NewServer()
	protos.RegisterLivetreamServiceServer(server, &rpc.Server{})

	log.Printf("server listening at %v", lis.Addr())
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
