package hive

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/sync/errgroup"

	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

const CollectionFaction = "faction"

type FactionRelationState uint

const (
	FactionRelationNeutral FactionRelationState = iota
	FactionRelationSendPeaceRequest
	FactionRelationPeace
	FactionRelationWar
)

type FactionMemberState uint

const (
	FactionMemberRequestJoin FactionMemberState = iota
	FactionMemberJoined
)

type FactionSector struct {
	SectorID bson.ObjectId `json:"sector_id" bson:"sector_id"`
	EntityID int64         `json:"entity_id" bson:"entity_id"`
}

type FactionRelation struct {
	FactionID bson.ObjectId        `json:"sector_id" bson:"sector_id"`
	Relation  FactionRelationState `json:"state" bson:"state"`
}

type FactionMember struct {
	SteamID  uint64             `json:"steam_id" bson:"steam_id"`
	State    FactionMemberState `json:"state" bson:"state"`
	IsLeader bool               `json:"is_leader" bson:"is_leader"`
}

type Faction struct {
	ID               bson.ObjectId     `json:"id" bson:"_id,omitempty"`
	HiveID           bson.ObjectId     `json:"-" bson:"hive_id"`
	Tag              string            `json:"tag" bson:"tag"`
	Name             string            `json:"name" bson:"name"`
	Description      string            `json:"description" bson:"description"`
	PrivateInfo      string            `json:"private_info" bson:"private_info"`
	AcceptHumans     bool              `json:"accept_humans" bson:"accept_humans"`
	FounderSteamID   uint64            `json:"founder_steam_id" bson:"founder_steam_id"`
	AutoAcceptMember bool              `json:"auto_accept_member" bson:"auto_accept_member"`
	AutoAcceptPeace  bool              `json:"auto_accept_peace" bson:"auto_accept_peace"`
	Relations        []FactionRelation `json:"relations" bson:"relations"`
	Members          []FactionMember   `json:"members" bson:"members"`
	Sectors          []FactionSector   `json:"sectors" bson:"sectors"`
}

func (s *System) GetFactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn := s.db.Copy()
	defer conn.Close()

	var factions []Faction
	err := conn.DB("torchhive").C(CollectionFaction).Find(bson.M{
		"hive_id": bson.ObjectIdHex(vars["hive_id"]),
	}).All(&factions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(factions); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *System) DeleteFactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	conn := s.db.Copy()
	defer conn.Close()

	_, err := conn.DB("torchhive").C(CollectionFaction).RemoveAll(bson.M{
		"hive_id": bson.ObjectIdHex(vars["hive_id"]),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *System) CreateFaction(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionCreated) error {
	conn := s.db.Copy()
	defer conn.Close()

	return conn.DB("torchhive").C(CollectionFaction).Insert(Faction{
		HiveID:       hiveID,
		Name:         event.Name,
		Tag:          event.Tag,
		Description:  event.Description,
		PrivateInfo:  event.PrivateInfo,
		AcceptHumans: event.AcceptHumans,
		Sectors: []FactionSector{
			{
				SectorID: sectorID,
				EntityID: event.FactionID,
			},
		},
	})
}

func (s *System) AddFactionSector(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionCreatedComplete) error {
	conn := s.db.Copy()
	defer conn.Close()

	return conn.DB("torchhive").C(CollectionFaction).Update(
		bson.M{
			"hive_id": hiveID,
			"tag":     event.Tag,
		},
		bson.M{
			"$push": bson.M{
				"sectors": bson.M{
					"sector_id": sectorID,
					"entity_id": event.FactionID,
				},
			},
		},
	)
}

func (s *System) GetFaction(hiveID bson.ObjectId, sectorID bson.ObjectId, factionID int64) (*Faction, error) {
	conn := s.db.Copy()
	defer conn.Close()

	var faction Faction
	err := conn.DB("torchhive").C(CollectionFaction).Find(bson.M{
		"hive_id": hiveID,
		"sectors": bson.M{
			"$elemMatch": bson.M{
				"sector_id": sectorID,
				"entity_id": factionID,
			},
		},
	}).One(&faction)
	if err != nil {
		return nil, err
	}

	return &faction, nil
}

func (s *System) EditFaction(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionEdited) error {
	conn := s.db.Copy()
	defer conn.Close()

	return conn.DB("torchhive").C(CollectionFaction).Update(
		bson.M{
			"hive_id": hiveID,
			"sectors": bson.M{
				"$elemMatch": bson.M{
					"sector_id": sectorID,
					"entity_id": event.FactionID,
				},
			},
		},
		bson.M{
			"$set": bson.M{
				"tag":          event.Tag,
				"name":         event.Name,
				"description":  event.Description,
				"private_info": event.PrivateInfo,
			},
		},
	)
}

func (s *System) ChangeAutoAccept(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionAutoAcceptChangeEvent) error {
	conn := s.db.Copy()
	defer conn.Close()

	return conn.DB("torchhive").C(CollectionFaction).Update(
		bson.M{
			"hive_id": hiveID,
			"sectors": bson.M{
				"$elemMatch": bson.M{
					"sector_id": sectorID,
					"entity_id": event.FactionID,
				},
			},
		},
		bson.M{
			"$set": bson.M{
				"auto_accept_member": event.AutoAcceptMember,
				"auto_accept_peace":  event.AutoAcceptPeace,
			},
		},
	)
}

func (s *System) SendPeaceRequest(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionPeaceWar) (*Faction, *Faction, error) {
	var fromFaction *Faction
	var toFaction *Faction
	errWg := errgroup.Group{}
	errWg.Go(func() error {
		var err error
		fromFaction, err = s.getFaction(hiveID, sectorID, event.FromFactionID)
		return err
	})
	errWg.Go(func() error {
		var err error
		toFaction, err = s.getFaction(hiveID, sectorID, event.ToFactionID)
		return err
	})
	err := errWg.Wait()
	if err != nil {
		return nil, nil, err
	}

	return fromFaction, toFaction, s.updateFactionRelation(FactionRelationSendPeaceRequest, fromFaction, toFaction)
}

func (s *System) CancelPeaceRequest(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionPeaceWar) (*Faction, *Faction, error) {
	var fromFaction *Faction
	var toFaction *Faction
	errWg := errgroup.Group{}
	errWg.Go(func() error {
		var err error
		fromFaction, err = s.getFaction(hiveID, sectorID, event.FromFactionID)
		return err
	})
	errWg.Go(func() error {
		var err error
		toFaction, err = s.getFaction(hiveID, sectorID, event.ToFactionID)
		return err
	})
	err := errWg.Wait()
	if err != nil {
		return nil, nil, err
	}

	// this should reset the relation back to the state of the to faction relation
	return fromFaction, toFaction, s.updateFactionRelation(FactionRelationNeutral, fromFaction, toFaction)
}

func (s *System) AcceptPeace(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionPeaceWar) (*Faction, *Faction, error) {
	var fromFaction *Faction
	var toFaction *Faction
	errWg := errgroup.Group{}
	errWg.Go(func() error {
		var err error
		fromFaction, err = s.getFaction(hiveID, sectorID, event.FromFactionID)
		return err
	})
	errWg.Go(func() error {
		var err error
		toFaction, err = s.getFaction(hiveID, sectorID, event.ToFactionID)
		return err
	})
	err := errWg.Wait()
	if err != nil {
		return nil, nil, err
	}

	return fromFaction, toFaction, s.updateFactionRelation(FactionRelationPeace, fromFaction, toFaction)
}

func (s *System) DeclareWar(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionPeaceWar) (*Faction, *Faction, error) {
	var fromFaction *Faction
	var toFaction *Faction
	errWg := errgroup.Group{}
	errWg.Go(func() error {
		var err error
		fromFaction, err = s.getFaction(hiveID, sectorID, event.FromFactionID)
		return err
	})
	errWg.Go(func() error {
		var err error
		toFaction, err = s.getFaction(hiveID, sectorID, event.ToFactionID)
		return err
	})
	err := errWg.Wait()
	if err != nil {
		return nil, nil, err
	}

	return fromFaction, toFaction, s.updateFactionRelation(FactionRelationWar, fromFaction, toFaction)
}

func (s *System) getFaction(hiveID bson.ObjectId, sectorID bson.ObjectId, entityID int64) (*Faction, error) {
	conn := s.db.Copy()
	defer conn.Close()

	var faction Faction
	err := conn.DB("torchhive").C(CollectionFaction).Find(bson.M{
		"hive_id": hiveID,
		"sectors": bson.M{
			"$elemMatch": bson.M{
				"sector_id": sectorID,
				"entity_id": entityID,
			},
		},
	}).One(&faction)
	if err != nil {
		return nil, err
	}

	return &faction, nil
}

func (s *System) updateFactionRelation(state FactionRelationState, fromFaction *Faction, toFaction *Faction) error {
	errWg := errgroup.Group{}
	errWg.Go(func() error {
		conn := s.db.Copy()
		defer conn.Close()

		found := false
		for _, v := range fromFaction.Relations {
			if v.FactionID != toFaction.ID {
				continue
			}

			found = true
			err := conn.DB("torchhive").C(CollectionFaction).Update(
				bson.M{
					"_id":                  fromFaction.ID,
					"relations.faction_id": toFaction.ID,
				},
				bson.M{
					"$set": bson.M{
						"relations.$.state": state,
					},
				},
			)
			if err != nil {
				return err
			}

			break
		}

		if found {
			return nil
		}

		return conn.DB("torchhive").C(CollectionFaction).Update(
			bson.M{
				"_id": fromFaction.ID,
			},
			bson.M{
				"$push": bson.M{
					"relations": bson.M{
						"faction_id": toFaction.ID,
						"state":      state,
					},
				},
			},
		)
	})
	errWg.Go(func() error {
		// leave old state if peace is requested
		if state == FactionRelationSendPeaceRequest {
			return nil
		}

		conn := s.db.Copy()
		defer conn.Close()

		found := false
		for _, v := range toFaction.Relations {
			if v.FactionID != fromFaction.ID {
				continue
			}

			found = true
			err := conn.DB("torchhive").C(CollectionFaction).Update(
				bson.M{
					"_id":                  fromFaction.ID,
					"relations.faction_id": toFaction.ID,
				},
				bson.M{
					"$set": bson.M{
						"relations.$.state": state,
					},
				},
			)
			if err != nil {
				return err
			}

			break
		}

		if found {
			return nil
		}

		return conn.DB("torchhive").C(CollectionFaction).Update(
			bson.M{
				"_id": fromFaction.ID,
			},
			bson.M{
				"$push": bson.M{
					"relations": bson.M{
						"faction_id": toFaction.ID,
						"state":      state,
					},
				},
			},
		)
	})

	return errWg.Wait()
}

func (s *System) MemberSendJoin(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionMember) (*Faction, error) {
	faction, err := s.getFaction(hiveID, sectorID, event.FactionID)
	if err != nil {
		return nil, err
	}

	for _, v := range faction.Members {
		if v.SteamID == event.PlayerSteamID {
			return nil, fmt.Errorf("steam id %d want to join in faction %s but exists", event.PlayerSteamID, faction.ID.Hex())
		}
	}

	conn := s.db.Copy()
	defer conn.Close()

	err = conn.DB("torchhive").C(CollectionFaction).Update(
		bson.M{
			"_id": faction.ID,
		},
		bson.M{
			"$push": bson.M{
				"members": bson.M{
					"steam_id": event.PlayerSteamID,
					"state":    FactionMemberRequestJoin,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return faction, nil
}

func (s *System) MemberLeave(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionMember) (*Faction, error) {
	faction, err := s.getFaction(hiveID, sectorID, event.FactionID)
	if err != nil {
		return nil, err
	}

	found := false
	for _, v := range faction.Members {
		if v.SteamID == event.PlayerSteamID {
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("steam id %d want to leave in faction %s but not exists", event.PlayerSteamID, faction.ID.Hex())
	}

	conn := s.db.Copy()
	defer conn.Close()

	err = conn.DB("torchhive").C(CollectionFaction).Update(
		bson.M{
			"_id": faction.ID,
		},
		bson.M{
			"$pull": bson.M{
				"members": bson.M{
					"steam_id": event.PlayerSteamID,
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return faction, nil
}

func (s *System) MemberAcceptJoin(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionMember) (*Faction, error) {
	faction, err := s.getFaction(hiveID, sectorID, event.FactionID)
	if err != nil {
		return nil, err
	}

	found := false
	for _, v := range faction.Members {
		if v.SteamID == event.PlayerSteamID {
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("steam id %d want to accept join in faction %s but not exists", event.PlayerSteamID, faction.ID.Hex())
	}

	conn := s.db.Copy()
	defer conn.Close()

	err = conn.DB("torchhive").C(CollectionFaction).Update(
		bson.M{
			"_id":              faction.ID,
			"members.steam_id": event.PlayerSteamID,
		},
		bson.M{
			"$set": bson.M{
				"members.$.steam_id": FactionMemberJoined,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return faction, nil
}

func (s *System) MemberPromoteDemote(hiveID bson.ObjectId, sectorID bson.ObjectId, event EventFactionMember, promote bool) (*Faction, error) {
	faction, err := s.getFaction(hiveID, sectorID, event.FactionID)
	if err != nil {
		return nil, err
	}

	found := false
	for _, v := range faction.Members {
		if v.SteamID == event.PlayerSteamID {
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("steam id %d want to accept join in faction %s but not exists", event.PlayerSteamID, faction.ID.Hex())
	}

	conn := s.db.Copy()
	defer conn.Close()

	err = conn.DB("torchhive").C(CollectionFaction).Update(
		bson.M{
			"_id":              faction.ID,
			"members.steam_id": event.PlayerSteamID,
		},
		bson.M{
			"$set": bson.M{
				"members.$.is_leader": promote,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return faction, nil
}
