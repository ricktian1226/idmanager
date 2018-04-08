package common

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

const (
	ID_TYPE_USER = iota //用户标识
	ID_TYPE_MAX
)

//@ Title id管理器定义
type IdManager struct {
	beginTimeMilliSecs int64 //开始时间毫秒数
	dividers           [ID_TYPE_MAX]*idDivider
}

func NewIdManager() (mgr *IdManager) {

	mgr = &IdManager{}
	for i, _ := range mgr.dividers {
		mgr.dividers[i] = NewIdDivider()
	}
	return
}

//@ Title 初始化id管理器
func (this *IdManager) Init() (err error) {
	beginTime, err := time.Parse(GO_BEGIN_TIME, ID_BEGIN_TIME)
	if err != nil {
		return
	}
	this.beginTimeMilliSecs = beginTime.UnixNano() / int64(time.Millisecond)
	return
}

const (
	GO_BEGIN_TIME = "2006-01-02 15:04:05" //go开始时间
	ID_BEGIN_TIME = "2018-01-01 00:00:00" //ID开始时间
)

//@ Title snowflake id分配器定义
type idDivider struct {
	lastTime int64 //上次相对毫秒数
	curIndex int64 //当前已分配的id计数
	lock     *sync.Mutex
}

func NewIdDivider() *idDivider {
	return &idDivider{
		lock: &sync.Mutex{},
	}
}

const (
	BITS_TIMESTAMP = 42
	BITS_REGION_ID = 4
	BITS_SERVER_ID = 10
	BITS_ID        = 8

	BITS_SHIFT_TIMESTAMP = BITS_REGION_ID + BITS_SERVER_ID + BITS_ID
	BITS_SHIFT_REGION_ID = BITS_SERVER_ID + BITS_ID
	BITS_SHIFT_SERVER_ID = BITS_ID
	BITS_SHIFT_ID        = 0
)

var (
	ID_MASK_TIMESTAMP = uint64(math.Pow(2, BITS_TIMESTAMP)-1) << BITS_SHIFT_TIMESTAMP
	ID_MASK_REGION    = uint64(math.Pow(2, BITS_REGION_ID)-1) << BITS_SHIFT_REGION_ID
	ID_MASK_SERVER    = uint64(math.Pow(2, BITS_SERVER_ID)-1) << BITS_SHIFT_SERVER_ID
	ID_MASK_ID        = uint64(math.Pow(2, BITS_ID)-1) << BITS_SHIFT_ID
)

//@ Title 生成唯一标识
//格式如下：
// |      timestamp  (42bits,可以使用139年) |  region id (4bits，支持最多16个区)  |  server id(10bits，每个区支持1024个节点)  |    id index  (8bits，每毫秒最多分配256个uid)     |
// |             相对时间，单位：毫秒        |               区域标识             |             服务器标识                   |                  id 计数器                      |
func (this *idDivider) GenUid() (err error, uid uint64) {
	//校验region id合法性
	if Cursvr.regionId > int64(math.Pow(2, BITS_SHIFT_REGION_ID)-1) ||
		Cursvr.regionId < 0 {
		//
		LOG_FUNC_ERROR("region id(%d) 非法，合法范围 (0, %d]", Cursvr.regionId, int64(math.Pow(2, BITS_SHIFT_REGION_ID)-1))
		return errors.New(fmt.Sprintf("region id(%d) 非法，合法范围 (0, %d]", Cursvr.regionId, int64(math.Pow(2, BITS_SHIFT_REGION_ID)-1))), 0
	}

	//校验server id合法性

	this.lock.Lock()
	defer this.lock.Unlock()

	timestampPart := (time.Now().UnixNano()/int64(time.Millisecond) - mgr.beginTimeMilliSecs) << BITS_SHIFT_TIMESTAMP
	regionPart := Cursvr.regionId << BITS_SHIFT_REGION_ID
	serverPart := Cursvr.serverId << BITS_SHIFT_SERVER_ID

	if this.lastTime != timestampPart {
		this.curIndex = 0
		this.lastTime = timestampPart
	} else {
		this.curIndex++
		//校验是否越界
	}

	return nil, uint64(timestampPart | regionPart | serverPart | this.curIndex)
}

func ID_Timestamp(timestamp int64) string {
	tm := time.Unix(int64((timestamp+mgr.beginTimeMilliSecs)/1000), 0)
	return tm.Format("2006-01-02 03:04:05 PM\n")
}

var mgr *IdManager

func ID_MANAGER_INIT() error {
	mgr = NewIdManager()
	return mgr.Init()
}

func ID_MANAGER_GEN(idType int) (err error, id uint64) {
	if idType < 0 || idType > ID_TYPE_MAX {
		LOG_FUNC_ERROR("错误的id类型 %d", idType)
		return errors.New(fmt.Sprintf("错误的id类型 %d", idType)), 0
	}

	return mgr.dividers[idType].GenUid()
}
