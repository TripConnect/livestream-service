package rpc

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/esdsl"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/sortorder"
	"github.com/gocql/gocql"
	"github.com/tripconnect/go-common-utils/common"
	pb "github.com/tripconnect/go-proto-lib/protos"
	"github.com/tripconnect/livestream-service/consts"
	"github.com/tripconnect/livestream-service/model"
)

const HLS_LINK_BASE = "/livestreams"

func (s *Server) CreateLivestream(ctx context.Context, req *pb.CreateLivestreamRequest) (*pb.Livestream, error) {
	livestreamId := gocql.MustRandomUUID()
	livestream := model.LivestreamEntity{
		Id:        livestreamId,
		Title:     req.GetTitle(),
		Thumbnail: req.GetThumbnail(),
		HlsLink:   HLS_LINK_BASE + "/" + livestreamId.String() + "/index.m3u8",
		Status:    model.CREATED.Int(),
		CreatedAt: time.Now(),
	}
	insertErr := model.LivestreamRepository.Insert(livestream)

	if insertErr != nil {
		log.Fatalf("Failed to insert livestream: %v", insertErr)
		return nil, insertErr
	}

	conversationDoc := model.NewLivestreamDoc(livestream)
	common.ElasticsearchClient.
		Index(consts.LivestreamIndex).
		Id(conversationDoc.Id.String()).
		Request(&conversationDoc).
		Do(ctx)

	livestreamPb := model.NewLivestreamPb(livestream)

	return &livestreamPb, nil
}

func (s *Server) FindLivestream(ctx context.Context, req *pb.FindLivestreamRequest) (*pb.Livestream, error) {
	livestream, err := model.LivestreamRepository.Get(req.GetLivestreamId())
	if err != nil {
		return nil, err
	}
	livestreamPb := model.NewLivestreamPb(*livestream.(*model.LivestreamEntity))
	return &livestreamPb, nil
}

func (s *Server) SearchLivestream(ctx context.Context, req *pb.SearchLivestreamsRequest) (*pb.Livestreams, error) {

	musts := []types.QueryVariant{
		esdsl.NewMatchAllQuery(),
	}

	pagerNumber := int(req.GetPageNumber())
	pageSize := int(req.GetPageSize())

	if req.GetStatus() != "" {
		status := model.FindLivestreamStatus(req.GetStatus())
		musts = append(musts, esdsl.NewMatchPhraseQuery("status", strconv.Itoa(int(status))))
	}

	if req.GetTerm() != "" {
		searchTerm := req.GetTerm()
		musts = append(musts, esdsl.NewWildcardQuery("title", searchTerm))
	}

	esQuery := esdsl.NewBoolQuery().
		Must(musts...)

	esResp, esErr := common.ElasticsearchClient.Search().
		Index(consts.LivestreamIndex).
		Query(esQuery).
		Sort(esdsl.NewSortOptions().AddSortOption("created_at", esdsl.NewFieldSort(sortorder.Desc))).
		From(pagerNumber * pageSize).
		Size(pageSize).
		Do(ctx)

	if esErr != nil {
		return nil, esErr
	}

	livestreamDocs := common.GetResponseDocs[model.LivestreamDocument](esResp)

	entities := []model.LivestreamEntity{}
	for _, doc := range livestreamDocs {
		if entity, err := model.LivestreamRepository.Get(doc.Id); err == nil {
			entities = append(entities, *entity.(*model.LivestreamEntity))
		} else {
			fmt.Printf("Error: %v", err)
		}
	}

	var livestreamPbs []*pb.Livestream
	for _, livestreamEntity := range entities {
		conversation := model.NewLivestreamPb(livestreamEntity)
		livestreamPbs = append(livestreamPbs, &conversation)
	}

	result := &pb.Livestreams{Livestreams: livestreamPbs}
	return result, nil
}
