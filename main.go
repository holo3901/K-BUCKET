package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"math/rand"
)

// Peer 代表一个节点
type Peer struct {
	ID         string           // 节点 ID,字符串表示
	KBuckets   [160]*Bucket     // Kademlia K 桶,共 160 个桶
	KnownPeers map[string]*Peer // 已知的对等节点
	keys       map[string][]byte
}

// Bucket 代表一个桶
type Bucket struct {
	Nodes []*Node // 桶中的节点
}

// Node 代表一个节点
type Node struct {
	ID      string // 节点 ID
	Address string // 节点地址
}

var nodesMap map[string]*Peer

var SetNodes []string
var GetNodes []string

// NewPeer 创建一个新的 Peer
func NewPeer(id string) *Peer {
	peer := new(Peer)
	peer.ID = id
	for i := 0; i < 160; i++ {
		peer.KBuckets[i] = new(Bucket)
	}
	peer.KnownPeers = make(map[string]*Peer) //新
	peer.keys = make(map[string][]byte)
	return peer
}

// InsertNode 将节点插入到对应桶中
func (p *Peer) InsertNode(nodeID string) {
	distance := Distance(p.ID, nodeID)               // 计算两个节点的距离，改
	bucketIndex := int(math.Log2(float64(distance))) // 求桶索引，改
	bucket := p.KBuckets[bucketIndex]
	// 如果桶为空,直接插入
	if len(bucket.Nodes) == 0 {
		bucket.Nodes = append(bucket.Nodes, &Node{ID: nodeID})
		return
	}
	// 如果节点已经在桶中,更新节点位置
	for i, node := range bucket.Nodes {
		if node.ID == nodeID {
			bucket.Nodes = append(bucket.Nodes[:i], bucket.Nodes[i+1:]...)
			bucket.Nodes = append([]*Node{&Node{ID: nodeID}}, bucket.Nodes...)
			return
		}
	}
	// 如果桶已满,随机替换一个节点
	if len(bucket.Nodes) >= 3 {
		r := rand.Intn(len(bucket.Nodes))
		bucket.Nodes[r] = &Node{ID: nodeID}
		return
	}
	// 桶未满,直接插入
	bucket.Nodes = append(bucket.Nodes, &Node{ID: nodeID})
}

// PrintBucketContents 打印每个桶中的节点
func (p *Peer) PrintBucketContents() {
	fmt.Println("----------------------", p.ID, "-----------------------------------")
	for i, bucket := range p.KBuckets {
		if len(bucket.Nodes) > 0 {
			fmt.Printf("Bucket %d: ", i)
			for _, node := range bucket.Nodes {
				fmt.Printf("%s ", node.ID)
			}
			fmt.Println()
		}
	}
}

// FindNode 查找节点
func (p *Peer) FindNode(nodeID string) bool {
	// 先插入节点
	p.InsertNode(nodeID)
	// 查找自己的桶是否有此节点
	for _, bucket := range p.KBuckets {
		for _, node := range bucket.Nodes {
			if node.ID == nodeID {
				return true
			}
		}
	}
	// 没有找到,随机选择两个节点查询
	distance := Distance(p.ID, nodeID)
	bucketIndex := int(math.Log2(float64(distance))) //改
	bucket := p.KBuckets[bucketIndex]
	var peers []*Peer
	for i := 0; i < 2 && i < len(bucket.Nodes); i++ {
		peer, ok := p.KnownPeers[bucket.Nodes[i].ID]
		if ok {
			peers = append(peers, peer)
		}
	}
	// 向选择的节点查询
	for _, peer := range peers {
		found := peer.FindNode(nodeID)
		if found {
			return true
		}
	}
	return false
}
func (p *Peer) Broadcast(newPeer *Peer) { //new
	// 将新节点广播给已知节点
	knownPeersCopy := make(map[string]*Peer)
	for k, v := range p.KnownPeers {
		knownPeersCopy[k] = v
	}
	for _, peer := range knownPeersCopy {
		peer.InsertNode(newPeer.ID)
	}
}
func (p *Peer) SetValue(key string, value []byte) bool {
	hash := sha1.Sum(value)
	hash_str := hex.EncodeToString(hash[:])
	if key != hash_str {
		return false
	}
	if p.keys[key] != nil {
		return true
	}

	//将内容存入自己的节点中
	p.keys[key] = value

	return true
}

func (s *Peer) GetValue(key string) []byte {
	if s.keys[key] != nil {
		hash := sha1.Sum(s.keys[key])

		hash_str := hex.EncodeToString(hash[:])
		if key != hash_str {
			return nil
		}

		return s.keys[key]
	}

	//获取到最近的桶
	return nil
}

func isUpdated(targetValue string, nodes []string, compare string) int {
	targetBinary := new(big.Int)
	targetBinary.SetString(targetValue, 2)
	compareBinary := new(big.Int)
	compareBinary.SetString(compare, 2)
	minValue := compareGetMin(targetValue, nodes[0], nodes[1])
	if minValue == nodes[0] {
		maxValueBinary := new(big.Int)
		maxValueBinary.SetString(nodes[1], 2)
		resultMaxValue := new(big.Int)
		resultMaxValue.Xor(targetBinary, maxValueBinary)
		resultCom := new(big.Int)
		resultCom.Xor(targetBinary, compareBinary)
		if resultCom.Cmp(resultMaxValue) < 0 {
			return 1
		}
	} else {
		maxValueBinary := new(big.Int)
		maxValueBinary.SetString(nodes[0], 2)
		resultMaxValue := new(big.Int)
		resultMaxValue.Xor(targetBinary, maxValueBinary)
		resultCom := new(big.Int)
		resultCom.Xor(targetBinary, compareBinary)
		if resultCom.Cmp(resultMaxValue) < 0 {
			return 0
		}
	}
	return -1
}

func checkLen(len int) (int, int) {
	if len > 2 {
		return GetRandom2()
	} else if len == 2 {
		return 0, 1
	} else {
		return 0, -1
	}
}
func compareGetMin(targetValue, value1, value2 string) string {
	num := new(big.Int)
	num1 := new(big.Int)
	num2 := new(big.Int)
	num.SetString(targetValue, 2)
	num1.SetString(value1, 2)
	num2.SetString(value2, 2)
	//计算出距离
	result1 := new(big.Int)
	result1.Xor(num, num1)
	result2 := new(big.Int)
	result2.Xor(num, num2)

	if result1.Cmp(result2) < 0 {
		return value1
	} else {
		return value2
	}
}
func GetRandom2() (int, int) {
	var nums [2]int
	// 随机生成两个不重复的整数
	for i := range nums {
		num := rand.Intn(3)
		nums[i] = int(num)
	}
	for nums[0] == nums[1] {
		num := rand.Intn(3)
		nums[1] = int(num)
	}
	return nums[0], nums[1]
}
func main() {
	// 初始化 5 个节点
	nodesMap = make(map[string]*Peer)
	peers := make(map[string]*Peer)
	x := make([]string, 0, 0)
	for i := 0; i < 5; i++ {
		peerID := GenerateID()
		peers[peerID] = NewPeer(peerID)
		peers[peerID].keys = make(map[string][]byte)
		x = append(x, peerID)
	}
	// 生成 200 个新节点,广播并加入网络
	for i := 0; i < 100; i++ {
		peerID := GenerateID()
		newPeer := NewPeer(peerID)
		peers[peerID] = newPeer
		nodesMap[peerID] = newPeer
		newPeer.keys = make(map[string][]byte)
		x = append(x, peerID)
	}
	for id1 := range peers {
		for id2 := range peers {
			if id1 != id2 {
				peers[id1].KnownPeers[id2] = peers[id2]
			}
		}
	}
	for a := 0; a < 100; a++ {
		// 随机选择一个节点广播
		rid := rand.Intn(len(peers))
		for peerIDs := range peers {
			if rid == 0 {
				peers[peerIDs].Broadcast(peers[peerIDs])
				break
			}
			rid--
		}
	}
	// 打印每个节点的桶
	for _, peer := range peers {
		peer.PrintBucketContents()
	}

	var strs []string
	var hashs []string
	for i := 0; i < 200; i++ {
		length := 8 // 随机生成长度为8的字符串
		bytes := make([]byte, length)
		rand.Read(bytes) // 从随机源中读取指定长度的随机字节序列
		str := base64.URLEncoding.EncodeToString(bytes)
		strs = append(strs, str)
		hash := sha1.Sum([]byte(str))
		hash_str := hex.EncodeToString(hash[:])
		hashs = append(hashs, hash_str)
	}

	//寻找一个随机节点
	num := rand.Intn(99)
	for i, v := range strs {
		hashInverse := inverse(hashs[i])
		if len(SetNodes) == 0 {
			SetNodes = append(SetNodes, hashInverse, hashInverse)
		} else {
			SetNodes[0] = hashInverse
			SetNodes[1] = hashInverse
		}
		peers[x[num]].SetValue(hashs[i], []byte(v))
	}

	//生成100个随机数
	var nums []int
	var isExist [200]bool
	for len(nums) <= 100 {
		num1 := rand.Intn(200)
		if !isExist[num1] {
			nums = append(nums, int(num1))
			isExist[num1] = true
		}
	}
	for _, v := range nums {
		num2 := rand.Intn(99)
		hashInverse := inverse(hashs[v])
		if len(GetNodes) == 0 {
			GetNodes = append(GetNodes, hashInverse, hashInverse)
		} else {
			GetNodes[0] = hashInverse
			GetNodes[1] = hashInverse
		}
		for i, _ := range peers[x[num2]].keys {
			value := peers[x[num2]].GetValue(i)
			fmt.Println("key is", hashs[v])
			if value != nil {
				fmt.Println("value is ", string(value))
			} else {
				fmt.Println("Can't find value")
			}
			fmt.Println("----------------------------------")
		}
	}
}
func Distance(id1, id2 string) int {
	var xor big.Int
	b1 := new(big.Int).SetBytes([]byte(id1))
	b2 := new(big.Int).SetBytes([]byte(id2))
	xor.Xor(b1, b2)
	return int(xor.Uint64())
}
func GenerateID() string {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}
func inverse(value string) string {
	byteArray := []byte(value)
	for i, v := range byteArray {
		if v == '0' {
			byteArray[i] = '1'
		} else {
			byteArray[i] = '0'
		}
	}
	return string(byteArray)
}
