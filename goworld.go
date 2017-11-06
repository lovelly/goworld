package goworld

import (
	"github.com/lovelly/goworld/components/game"
	"github.com/lovelly/goworld/engine/common"
	"github.com/lovelly/goworld/engine/entity"
	"github.com/lovelly/goworld/engine/kvdb"
	"github.com/lovelly/goworld/engine/post"
	"github.com/lovelly/goworld/engine/storage"
)

// Run runs the server endless loop
//
// This is the main routine for the server and all entity logic,
// and this function never quit
func Run(delegate game.IGameDelegate) {
	game.Run(delegate)
}

// RegisterEntity registers the entity type so that entities can be created or loaded
//
// returns the entity type description object which can be used to define more properties
// of entity type
func RegisterEntity(typeName string, entityPtr interface{}, isPersistent, useAOI bool) *entity.EntityTypeDesc {
	return entity.RegisterEntity(typeName, entityPtr, isPersistent, useAOI)
}

// CreateSpaceAnywhere creates a space with specified kind in any game server
func CreateSpaceAnywhere(kind int) {
	entity.CreateSpaceAnywhere(kind)
}

// CreateSpaceLocally creates a space with specified kind in the local game server
//
// returns the space EntityID
func CreateSpaceLocally(kind int) common.EntityID {
	return entity.CreateSpaceLocally(kind)
}

// CreateEntityLocally creates a entity on the local server
//
// returns EntityID
func CreateEntityLocally(typeName string) common.EntityID {
	return entity.CreateEntityLocally(typeName, nil, nil)
}

// CreateEntityAnywhere creates a entity on any server
func CreateEntityAnywhere(typeName string) {
	entity.CreateEntityAnywhere(typeName)
}

// LoadEntityAnywhere loads the specified entity from entity storage
func LoadEntityAnywhere(typeName string, entityID common.EntityID) {
	entity.LoadEntityAnywhere(typeName, entityID)
}

// GetServiceProviders get the set of EntityIDs that provides the specified service
func GetServiceProviders(serviceName string) entity.EntityIDSet {
	return entity.GetServiceProviders(serviceName)
}

// ListEntityIDs gets all saved entity ids in storage, may take long time and block the main routine
//
// returns result in callback
func ListEntityIDs(typeName string, callback storage.ListCallbackFunc) {
	storage.ListEntityIDs(typeName, callback)
}

// Exists checks if entityID exists in entity storage
//
// returns result in callback
func Exists(typeName string, entityID common.EntityID, callback storage.ExistsCallbackFunc) {
	storage.Exists(typeName, entityID, callback)
}

// GetEntity gets the entity by EntityID
func GetEntity(id common.EntityID) *entity.Entity {
	return entity.GetEntity(id)
}

// GetGameID gets the local server ID
//
// server ID is a uint16 number starts from 1, which should be different for each servers
// server ID is also in the game config section name of goworld.ini
func GetGameID() uint16 {
	return game.GetGameID()
}

// MapAttr creates a new MapAttr
func MapAttr() *entity.MapAttr {
	return entity.NewMapAttr()
}

// ListAttr creates a new ListAttr
func ListAttr() *entity.ListAttr {
	return entity.NewListAttr()
}

// RegisterSpace registers the space entity type.
//
// All spaces will be created as an instance of this type
func RegisterSpace(spacePtr interface{}) {
	entity.RegisterSpace(spacePtr)
}

// Entities gets all entities as an EntityMap (do not modify it!)
func Entities() entity.EntityMap {
	return entity.Entities()
}

// Post posts a callback to be executed
func Post(callback post.PostCallback) {
	post.Post(callback)
}

// GetKVDB gets value of key from KVDB
func GetKVDB(key string, callback kvdb.KVDBGetCallback) {
	kvdb.Get(key, callback)
}

// PutKVDB puts key-value to KVDB
func PutKVDB(key string, val string, callback kvdb.KVDBPutCallback) {
	kvdb.Put(key, val, callback)
}

// GetOrPut gets value of key from KVDB, if val not exists or is "", put key-value to KVDB.
func GetOrPutKVDB(key string, val string, callback kvdb.KVDBGetOrPutCallback) {
	kvdb.GetOrPut(key, val, callback)
}
