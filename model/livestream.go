package model

import (
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/esdsl"
	"github.com/gocql/gocql"
	"github.com/kristoiv/gocqltable"
	"github.com/kristoiv/gocqltable/recipes"
	pb "github.com/tripconnect/go-proto-lib/protos"
	"github.com/tripconnect/livestream-service/consts"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LivestreamStatus int

const (
	UNKNOWN LivestreamStatus = 0
	CREATED LivestreamStatus = 1
	READY   LivestreamStatus = 2
)

func FindLivestreamStatus(s string) LivestreamStatus {
	switch s {
	case "CREATED":
		return CREATED
	case "READY":
		return READY
	default:
		return UNKNOWN
	}
}

func (s LivestreamStatus) String() string {
	switch s {
	case CREATED:
		return "CREATED"
	case READY:
		return "READY"
	default:
		return "UNKNOWN"
	}
}

func (s LivestreamStatus) Int() int {
	return int(s)
}

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
	AddProperty("title", esdsl.NewKeywordProperty()).
	AddProperty("thumbnail", esdsl.NewKeywordProperty()).
	AddProperty("hls_link", esdsl.NewKeywordProperty()).
	AddProperty("status", esdsl.NewLongNumberProperty()).
	AddProperty("created_at", esdsl.NewLongNumberProperty())

func NewLivestreamDoc(entity LivestreamEntity) LivestreamDocument {
	return LivestreamDocument{
		Id:        entity.Id,
		Status:    int(entity.Status),
		Title:     entity.Title,
		Thumbnail: entity.Thumbnail,
		HlsLink:   entity.HlsLink,
		CreatedAt: int(entity.CreatedAt.UnixMilli()),
	}
}

func NewLivestreamPb(entity LivestreamEntity) pb.Livestream {
	return pb.Livestream{
		Id:         entity.Id.String(),
		Status:     LivestreamStatus(entity.Status).String(),
		Title:      entity.Title,
		Thumbnail:  entity.Thumbnail,
		HlsLink:    entity.HlsLink,
		CreateTime: timestamppb.New(entity.CreatedAt),
	}
}
