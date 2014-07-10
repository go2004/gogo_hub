/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 14-6-10
 * Time: 下午2:57
 * To change this template use File | Settings | File Templates.
 */
package main

import (
	"sync"
)

//全局变量
var (
	allUserSem    sync.RWMutex     // Used to synchronize access to all data structures in this var group
	allUser       [MAX_PLAYERS]*user // This array contains all socket
)

const (
	PlayerConnStateFree  = iota // The player has to Free
	PlayerConnStateLogin = iota // The player has to login
	PlayerConnStatePass  = iota // The player has to provide a password
	PlayerConnStateIn    = iota // The player is in the world
	PlayerConnStateDisc  = iota // The player is logged in but disconnected
)

//初始化连接数据
func InitGlobalUser() {
	var index uint32 = 0
	for i := 0; i < MAX_PLAYERS; i++ {
		up := new(user)
		up.id = index
		up.conn = nil
		up.state = PlayerConnStateFree
		allUser[i] = up
		index = index + 1
	}

	LogInfo(" InitGlobalUser is succes, Max_Users = ", MAX_PLAYERS)
}

//得到空闲用户
func GetFreeUser() (ok bool, pUser *user) {
	allUserSem.Lock()
	var index int
	for index = 0; index < MAX_PLAYERS; index++ {
		if allUser[index].state == PlayerConnStateFree {
			allUser[index].state = PlayerConnStateLogin
			break
		}
	}
	allUserSem.Unlock()
	if index >= MAX_PLAYERS {
		LogInfo(" Handle the case with too many players,index = ", index, ",MAX_PLAYERS =\n", MAX_PLAYERS)
		return false, nil
	}
	return true, allUser[index]
}

//释放用户数据
func FreeUser(pUser *user) (ok bool ) {
	//fmt.Printf(" FreeUser id = %d\n",pUser.id)
	allUserSem.Lock()
	pUser.state = ConnStateFree;
	allUserSem.Unlock()
	return true
}

func SaveAllPlayers() {
	allUserSem.RLock()
	for i := 0; i < MAX_PLAYERS; i++ {
		up := allUser[i]
		if up.state >= PlayerConnStatePass {
			up.forceSave = true
		}
	}
	allUserSem.RUnlock()
}

//设置连接序号到对应表中
func SetConn2User(pConn *connection, pUser *user) {
	if (pConn == nil || pUser == nil) {
		return
	}
	allSocketsSem.Lock()
	pConn.user = pUser
	allSocketsSem.Unlock()

	allUserSem.Lock()
	pUser.conn = pConn
	allUserSem.Unlock()
}
