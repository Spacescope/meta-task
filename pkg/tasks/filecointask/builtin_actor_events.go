package filecointask

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/Spacescore/observatory-task/pkg/models/filecoinmodel"
	"github.com/Spacescore/observatory-task/pkg/tasks/common"

	// "github.com/filecoin-project/go-state-types/cbor"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"
)

// RawActor extract raw actor
type BuiltInActorEvent struct {
}

func (r *BuiltInActorEvent) Name() string {
	return "builtin_actor_events"
}

func (r *BuiltInActorEvent) Model() interface{} {
	return new(filecoinmodel.BuiltinActorEvents)
}

// https://filecoinproject.slack.com/archives/C01V2SPTURL/p1711536308451839?thread_ts=1710320448.273839&cid=C01V2SPTURL
// https://docs.google.com/document/d/19OSO82VbfjGSx4QP-lg2OkuPHttChJyykd-HMTyJ5Mc/edit

var (
	fields  map[string][]types.ActorEventBlock
	convert map[string]string
)

type KVEvent struct {
	Key   string
	Value string
}

const (
	INT    = "int"
	STRING = "string"
	CID    = "cid"
	BIGINT = "bigint"
)

func init() {
	// ------------ fields ------------
	const (
		VerifierBalance    = "cHZlcmlmaWVyLWJhbGFuY2U="     // verifier-balance
		Allocation         = "amFsbG9jYXRpb24="             // allocation
		AllocationRemoved  = "cmFsbG9jYXRpb24tcmVtb3ZlZA==" // allocation-removed
		Claim              = "ZWNsYWlt"                     // claim
		ClaimUpdated       = "bWNsYWltLXVwZGF0ZWQ="         // claim-updated
		ClaimRemoved       = "bWNsYWltLXJlbW92ZWQ="         // claim-removed
		DealPublished      = "ZGVhbC1wdWJsaXNoZWQ="         // deal-published
		DealActivated      = "bmRlYWwtYWN0aXZhdGVk"         // deal-activated
		DealTerminated     = "b2RlYWwtdGVybWluYXRlZA=="     // deal-terminated
		DealCompleted      = "bmRlYWwtY29tcGxldGVk"         // deal-completed
		SectorPrecommitted = "c3NlY3Rvci1wcmVjb21taXR0ZWQ=" // sector-precommitted
		SectorActivated    = "cHNlY3Rvci1hY3RpdmF0ZWQ="     // sector-activated
		SectorUpdated      = "bnNlY3Rvci11cGRhdGVk"         // sector-updated
		SectorTerminated   = "cXNlY3Rvci10ZXJtaW5hdGVk"     // sector-terminated
	)

	verifierBalanceByte, _ := b64.StdEncoding.DecodeString(VerifierBalance)
	allocationByte, _ := b64.StdEncoding.DecodeString(Allocation)
	allocationRemovedByte, _ := b64.StdEncoding.DecodeString(AllocationRemoved)
	claimByte, _ := b64.StdEncoding.DecodeString(Claim)
	claimUpdatedByte, _ := b64.StdEncoding.DecodeString(ClaimUpdated)
	claimRemovedByte, _ := b64.StdEncoding.DecodeString(ClaimRemoved)
	dealPublishedByte, _ := b64.StdEncoding.DecodeString(DealPublished)
	dealActivatedByte, _ := b64.StdEncoding.DecodeString(DealActivated)
	dealTerminatedByte, _ := b64.StdEncoding.DecodeString(DealTerminated)
	dealCompletedByte, _ := b64.StdEncoding.DecodeString(DealCompleted)
	sectorPrecommittedByte, _ := b64.StdEncoding.DecodeString(SectorPrecommitted)
	sectorActivatedByte, _ := b64.StdEncoding.DecodeString(SectorActivated)
	sectorUpdatedByte, _ := b64.StdEncoding.DecodeString(SectorUpdated)
	sectorTerminatedByte, _ := b64.StdEncoding.DecodeString(SectorTerminated)

	fields = map[string][]types.ActorEventBlock{
		"$type": []types.ActorEventBlock{
			{81, verifierBalanceByte},    // verifier-balance
			{81, allocationByte},         // allocation
			{81, allocationRemovedByte},  // allocation-removed
			{81, claimByte},              // claim
			{81, claimUpdatedByte},       // claim-updated
			{81, claimRemovedByte},       // claim-removed
			{81, dealPublishedByte},      // deal-published
			{81, dealActivatedByte},      // deal-activated
			{81, dealTerminatedByte},     // deal-terminated
			{81, dealCompletedByte},      // deal-completed
			{81, sectorPrecommittedByte}, // sector-precommitted
			{81, sectorActivatedByte},    // sector-activated
			{81, sectorUpdatedByte},      // sector-updated
			{81, sectorTerminatedByte},   // sector-terminated
		},
	}

	// ------------ convert ------------
	// https://fips.filecoin.io/FIPS/fip-0083.html
	convert = map[string]string{
		"$type":        STRING,
		"verifier":     INT,
		"client":       INT,
		"balance":      BIGINT,
		"id":           INT,
		"provider":     INT,
		"piece-cid":    CID,
		"piece-size":   INT,
		"term-min":     INT,
		"term-max":     INT,
		"expiration":   INT,
		"term-start":   INT,
		"sector":       INT,
		"unsealed-cid": CID,
	}
}

func cborValueDecode(key string, value []byte) interface{} {
	var (
		resultSTR    string
		resultINT    int
		resultBIGINT types.BigInt
		resultCID    cid.Cid
		err          error
	)

	switch convert[key] {
	case STRING:
		err = cbor.Unmarshal(value, &resultSTR)
		if err != nil {
			log.Errorf("cbor.Unmarshal err: %v, key: %v", err, key)
			return nil
		}
		return resultSTR
	case INT:
		err = cbor.Unmarshal(value, &resultINT)
		if err != nil {
			log.Errorf("cbor.Unmarshal err: %v, key: %v", err, key)
			return nil
		}
		return resultINT
	case BIGINT:
		err = cbor.Unmarshal(value, &resultBIGINT)
		if err != nil {
			log.Errorf("cbor.Unmarshal err: %v, key: %v", err, key)
			return nil
		}
		return resultBIGINT
	case CID:
		err = cbor.Unmarshal(value, &resultCID)
		if err != nil {
			log.Errorf("cbor.Unmarshal err: %v, key: %v", err, key)
			return nil
		}
		return resultCID
	}

	return nil
}

func (r *BuiltInActorEvent) Run(ctx context.Context, tp *common.TaskParameters) error {
	tsKey := tp.AncestorTs.Key()
	filter := &types.ActorEventFilter{
		TipSetKey: &tsKey,
		Fields:    fields,
	}

	events, err := tp.Api.GetActorEventsRaw(ctx, filter)
	if err != nil {
		log.Errorf("GetActorEventsRaw[pTs: %v, pHeight: %v, cTs: %v, cHeight: %v] err: %v", tp.AncestorTs.String(), tp.AncestorTs.Height(), tp.CurrentTs.String(), tp.CurrentTs.Height(), err)
		return err
	}

	var eventsResult []*filecoinmodel.BuiltinActorEvents

	for evtIdx, event := range events {
		var eventsSlice []*KVEvent
		for entryIdx, e := range event.Entries {
			if e.Codec != 0x51 { // 81
				log.Warnf("Codec not equal to cbor, height: %v, evtIdx: %v, emitter: %v, entryIdx: %v, e.Codec: %v", tp.AncestorTs.Height(), evtIdx, event.Emitter.String(), entryIdx, e.Codec)
				continue
			}

			var kvEvent KVEvent
			kvEvent.Key = e.Key

			v := cborValueDecode(e.Key, e.Value)
			switch convert[e.Key] {
			case STRING:
				kvEvent.Value = v.(string)
			case INT:
				kvEvent.Value = strconv.Itoa(v.(int))
			case BIGINT:
				kvEvent.Value = v.(types.BigInt).String()
			case CID:
				if v != nil {
					kvEvent.Value = v.(string)
				}
			}
			eventsSlice = append(eventsSlice, &kvEvent)
		}

		re, _ := json.Marshal(eventsSlice)

		eventsResult = append(eventsResult, &filecoinmodel.BuiltinActorEvents{
			Height:     int64(tp.AncestorTs.Height()),
			Version:    tp.Version,
			MessageCid: event.MsgCid.String(),
			Emitter:    event.Emitter.String(),
			EventEntry: string(re),
		})
	}

	if err = common.InsertMany(ctx, new(filecoinmodel.BuiltinActorEvents), int64(tp.AncestorTs.Height()), tp.Version, &eventsResult); err != nil {
		log.Errorf("Sql Engine err: %v", err)
		return err
	}

	log.Infof("has been process %v builtin_actor_events", len(eventsResult))

	return nil
}
