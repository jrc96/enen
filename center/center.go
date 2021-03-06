package center

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"sync"
	"time"

	"enen/common/pb"

	"github.com/golang/protobuf/proto"
	"github.com/laonsx/gamelib/crypt"
	"github.com/laonsx/gamelib/pools"
	"github.com/laonsx/gamelib/rpc"
	"github.com/laonsx/gamelib/server"
	"github.com/sirupsen/logrus"
)

type CenterServer struct {
	mux   sync.RWMutex
	users map[uint64]*user
}

var centerServer *CenterServer

func NewLoginServer() *CenterServer {

	centerServer = new(CenterServer)
	centerServer.users = make(map[uint64]*user)

	return centerServer
}

func (center *CenterServer) Open(conn server.Conn) {

	ips := conn.RemoteAddr().String()

	logrus.Infof("new conn id=%d addr=%s", conn.Id(), ips)

	conn.SetMsgType(server.WS_MSG_STRING)
	conn.SetReadDeadline(time.Duration(10) * time.Second)

	chanMsg := make(chan []byte)
	go func() {

		defer func() {

			conn.Close()
			close(chanMsg)
		}()

		for {

			b, err := conn.Read()
			if err != nil {

				return
			}

			chanMsg <- b
		}
	}()

	center.start(chanMsg, conn)
}

func (center *CenterServer) Close(conn server.Conn) {

	logrus.Infof("conn closing id=%d", conn.Id())
}

func (center *CenterServer) start(chanMsg chan []byte, conn server.Conn) {

	challenge := crypt.Randomkey().String()
	err := conn.AsyncSend([]byte(challenge))

	//clientkey
	msg := <-chanMsg
	clikey := string(msg)

	//serverkey
	privatekey := crypt.Randomkey()

	//
	n := new(big.Int)
	P := new(big.Int)
	P.SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE65381FFFFFFFFFFFFFFFF", 16)
	G := new(big.Int)
	G.SetInt64(2)
	serverkey := n.Exp(G, privatekey, P)
	//

	//serverkey := crypt.DHExchange(privatekey)
	conn.AsyncSend([]byte(serverkey.String()))

	exchangekey := new(big.Int)
	exchangekey.SetString(string(clikey), 10)

	//
	nn := new(big.Int)

	secret := nn.Exp(exchangekey, privatekey, P)
	//

	//secret := crypt.DHSecret(privatekey, exchangekey)

	//验证双方密钥正确性
	bmac := <-chanMsg
	climac := string(bmac)
	mac := hmac.New(sha256.New, []byte(challenge))
	mac.Write([]byte(secret.String()))
	out := mac.Sum(nil)
	smac, _ := base64.StdEncoding.DecodeString(string(climac))
	if !hmac.Equal(smac, out) {

		conn.AsyncSend([]byte("err"))

		return
	}

	conn.AsyncSend([]byte("ok"))

	ret := make(map[string]interface{})
	response := func(ret map[string]interface{}) {

		jsondata, _ := json.Marshal(ret)
		conn.AsyncSend(jsondata)
	}

	binfo := <-chanMsg
	arr := strings.Split(string(binfo), ":")
	if len(arr) != 2 {

		ret["code"] = 2
		response(ret)

		return
	}

	token := arr[0]
	pf := arr[1]

	//验证token(auth服校验)
	openId, err := checkToken(token, pf)
	if err != nil {

		logrus.Warnf("checkToken err=%s", err.Error())

		ret["code"] = 3
		response(ret)

		return
	}

	//生成uid
	userId, err := getUserId(openId, pf)
	if err != nil {

		logrus.Warnf("getUserId err=%s token=%s openId=%s pf=%s", err.Error(), token, openId, pf)

		ret["code"] = 4
		response(ret)

		return
	}

	logrus.Infof("login token=%s pf=%s", token, pf)

	//通知gate
	req := pb.GateRequest{}
	req.Secret = secret.String()
	req.Uid = userId
	reqData, err := proto.Marshal(&req)
	if err != nil {

		logrus.Errorf("call gate marshal err=%s", err.Error())

		ret["code"] = 5
		response(ret)

		return
	}

	center.mux.Lock()
	oldUser := center.users[userId]
	center.mux.Unlock()

	var gateInfo *GateInfo

	if oldUser != nil {

		gateInfo = GateManager.getGateByName(oldUser.gate)
	} else {

		gateInfo = GateManager.getRandGateInfo()
	}

	_, err = rpc.Call(gateInfo.name, "GateService.Login", reqData, nil)
	if err != nil {

		logrus.Errorf("call gate err=%s", err.Error())

		ret["code"] = 6
		response(ret)

		return
	}

	user := new(user)
	user.ip = conn.RemoteAddr().String()
	user.secret = secret.String()
	user.uid = userId
	user.gate = gateInfo.name

	center.mux.Lock()
	center.users[user.uid] = user
	center.mux.Unlock()

	ret["code"] = 0
	ret["addr"] = gateInfo.addr
	ret["pfuid"] = openId
	ret["uid"] = userId

	response(ret)
}

func (center *CenterServer) delUser(uid uint64) {

	center.mux.Lock()
	defer center.mux.Unlock()

	delete(center.users, uid)
}

func (center *CenterServer) setOnline(uid uint64) error {

	center.mux.RLock()
	defer center.mux.RUnlock()

	if u, ok := center.users[uid]; ok {

		u.setOnline()

		return nil
	}

	return errors.New("no user")
}

func (center *CenterServer) setOffline(uid uint64) error {

	center.mux.RLock()
	defer center.mux.RUnlock()

	if u, ok := center.users[uid]; ok {

		u.setOffline()

		return nil
	}

	return errors.New("no user")
}

func (center *CenterServer) onlineList(uidlist []uint64) []bool {

	center.mux.RLock()
	defer center.mux.RUnlock()

	state := make([]bool, len(uidlist))
	for k, uid := range uidlist {

		if u, ok := center.users[uid]; ok {

			state[k] = u.online
		}
	}

	return state
}

func (center *CenterServer) getUser(uid uint64) (*user, error) {

	center.mux.RLock()
	defer center.mux.RUnlock()

	if u, ok := center.users[uid]; ok {

		return u, nil
	}

	return nil, errors.New("no user")
}

type user struct {
	uid    uint64
	online bool
	ip     string
	secret string
	gate   string
}

func (u *user) setOffline() {

	u.online = false
}

func (u *user) setOnline() {

	u.online = true
}

type stGameUserInfo struct {
	Code int                `json:"code"`
	Data stGameUserInfoData `json:"data"`
}

type stGameUserInfoData struct {
	Uid string `json:"uid"`
}

func checkToken(token string, pf string) (string, error) {

	if pf == "test" {

		return token, nil
	}

	return "", nil
}

var idPool = pools.NewIdPool(10001)

func getUserId(openid string, pf string) (uint64, error) {

	//todo getUserid

	return uint64(idPool.Get()), nil
}
