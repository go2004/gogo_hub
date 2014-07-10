/**
 * Created with IntelliJ IDEA.
 * User: Administrator
 * Date: 14-6-10
 * Time: 下午2:53
 * To change this template use File | Settings | File Templates.
 */
package main

type user struct {
	id					   	uint32 //
	conn 					*connection
	state                        uint8  // The TCP/IP connectin to the player.
	flags                        uint32 // Bit mapped flags that the client always have to know about. See UserFlag* in client_prot.
	forceSave				    bool
}

func (up *user) loginAck() {

}

