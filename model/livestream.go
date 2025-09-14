package model

import (
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/esdsl"
	"github.com/gocql/gocql"
	"github.com/kristoiv/gocqltable"
	"github.com/kristoiv/gocqltable/recipes"
	"github.com/tripconnect/livestream-service/consts"
)

type LivestreamEntity struct {
	Id        gocql.UUID `cql:"id"`
	Title     string     `cql:"title"`
	Thumbnail string     `cql:"thumbnail"`
	HlsLink   string     `cql:"hls_link"`
	Status    int        `cql:"status"`
	CreatedAt time.Time  `cql:"created_at"`
}

type LivestreamDocument struct {
	Id        gocql.UUID `json:"id"`
	Title     string     `json:"title"`
	Thumbnail string     `json:"thumbnail"`
	HlsLink   string     `json:"hls_link"`
	Status    int        `json:"status"`
	CreatedAt int        `json:"created_at"`
}

var LivestreamRepository = struct {
	recipes.CRUD
}{
	recipes.CRUD{
		TableInterface: gocqltable.NewKeyspace(consts.KeySpace).NewTable(
			consts.LivestreamTableName,
			[]string{"id"},
			nil,
			LivestreamEntity{},
		),
	},
}

var LivestreamDocumentMappings = esdsl.NewTypeMapping().
	AddProperty("id", esdsl.NewKeywordProperty()).
	AddProperty("conversation_id", esdsl.NewKeywordProperty()).
	AddProperty("from_user_id", esdsl.NewKeywordProperty()).
	AddProperty("content", esdsl.NewKeywordProperty()).
	AddProperty("sent_time", esdsl.NewLongNumberProperty()).
	AddProperty("created_at", esdsl.NewLongNumberProperty())
