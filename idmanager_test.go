package common

import (
	"testing"
)

func TestIdManager_GenUid(t *testing.T) {
	Cursvr = NewSvrBase()
	ID_MANAGER_INIT()
	var uids []uint64
	for i := 0; i < 2; i++ {
		err, uid := ID_MANAGER_GEN(ID_TYPE_USER)
		if err != nil {
			t.Fatalf("ID_MANAGER_GEN failed %v", err)
		}
		uids = append(uids, uid)
	}

	//uid := uint64(21795382048587776)
	for _, uid := range uids {
		timestamp := (uid & ID_MASK_TIMESTAMP) >> BITS_SHIFT_TIMESTAMP
		t.Log(ID_Timestamp(int64(timestamp)))
		regionId := (uid & ID_MASK_REGION) >> BITS_SHIFT_REGION_ID
		if regionId != 0 {
			t.Fatalf("region(%d) is error!", regionId)
		}

		serverId := (uid & ID_MASK_SERVER) >> BITS_SHIFT_SERVER_ID
		if serverId != 0 {
			t.Fatalf("server(%d) id is error!", serverId)
		}

		id := (uid & ID_MASK_ID) >> BITS_SHIFT_ID
		if id != 0 && id != 1 {
			t.Fatalf("id(%d) is error!", id)
		}
	}
}
