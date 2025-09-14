package rpc

import (
	"context"
	"log"

	"github.com/gocql/gocql"
	"github.com/tripconnect/go-common-utils/common"
	pb "github.com/tripconnect/go-proto-lib/protos"
	"github.com/tripconnect/livestream-service/consts"
	"github.com/tripconnect/livestream-service/model"
)

const HLS_LINK_BASE = "/livestreams/"

func (s *Server) CreateLivestream(ctx context.Context, req *pb.CreateLivestreamRequest) (*pb.Livestream, error) {
	livestreamId := gocql.MustRandomUUID()
	livestream := model.LivestreamEntity{
		Id:        livestreamId,
		Title:     req.GetTitle(),
		Thumbnail: req.GetThumbnail(),
		HlsLink:   HLS_LINK_BASE + "/" + livestreamId.String() + "/index.m3u8",
		Status:    model.CREATED,
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
