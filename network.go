package main

import (
	"context"
	"log"
	"strconv"
	pv "test-verifier/proto_verify"
	"time"
)

type verifySrv struct {
	pv.UnimplementedCommitteeServiceServer
}

func newVerifyService() *verifySrv {
	return &verifySrv{}
}

func (v *verifySrv) RequestCommittee(_ context.Context, cr *pv.CommitteeRequest) (*pv.CommitteeInfo, error) {
	log.Println("Received committee request")
	candidates := cr.Candidates

	var committeeNodes []string
	for _, cand := range candidates {
		committeeNodes = append(committeeNodes, cand.NodeId)
	}
	leader := candidates[0].NodeId

	return &pv.CommitteeInfo{
		ChannelId:      "<test>",
		MemberIds:      committeeNodes,
		LeaderMemberId: leader,
		Timestamp:      strconv.FormatInt(time.Now().Unix(), 10),
	}, nil
}
