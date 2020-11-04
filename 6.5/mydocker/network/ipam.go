package network

import (
	"net"
	"os"
	"path"
	"strings"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
)

const ipamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"

type IPAM struct {
	SubnetAllocatorPath string
	Subnets *map[string]string
}

var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

func (ipam *IPAM) load() error {
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		log.Errorf("Error dump allocation info, %v", err)
		return err
	}
	return nil
}

func (ipam *IPAM) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigFileDir, 0644)
		} else {
			return err
		}
	}
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC | os.O_WRONLY | os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}

	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}

	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}

	return nil
}

func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 存放网段中地址分配信息的数组
	ipam.Subnets = &map[string]string{}

	// 从文件中加载已经分配的网段信息
	err = ipam.load()
	if err != nil {
		log.Errorf("Error dump allocation info, %v", err)
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	//net.IPNet.Mask.Size （）函数会返回网段的子网掩码的总长度和网段前面的固定位的长度
	//比如“ 127.0.0.0/8 ”网段的子网掩码是“ 255.0.0.0”
	//那么 subnet Mask.Size ｛）的返回值就是前面 255 所对应的位数和总位数，即8和24
	one, size := subnet.Mask.Size()

	//如果之前没有分配过这个网段，则初始化网段的分配配置
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1 << uint8(size - one))
	}

	//／遍历网段的位图数组
	for c := range((*ipam.Subnets)[subnet.String()]) {
		//找到数组中为“0”的项和数组序号，即可以分配的 IP
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			//／设置这个为“ 0”的序号值为“ l1” 即分配这个 IP
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			///Go 的字符串，创建之后就不能修改 所以通过转换成 byte 数组，修改后再转换成字符串赋值
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			//／这里的 IP 为初始ip，比如对于网段 192 168.0.0/16 ，这里就是 192.168.0.0
			ip = subnet.IP

			/*
			通过网段的 IP 与上面的偏移相加计算出分配的 IP 地址，由于 IP 地址是 uint 个数组，
			商要通过数组中的每 项加所窝耍的值，比如网段是 172.16.0.0/12 ，数组序号是 65555.
			那么在［172,16,0,0］上依次加［ uint8(65555 >> 24 ）、 uint8(65555 >> 16 ）、
			uint8(65555 >> 8）、 uint8 (65555 >> 0) ，即［0, 1 , 0 , 19 ），那么获得的 IP就是172.17.0.19.
			*/
			for t := uint(4); t > 0; t-=1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3]+=1
			break
		}
	}

	ipam.dump()
	return
}

func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	err := ipam.load()
	if err != nil {
		log.Errorf("Error dump allocation info, %v", err)
	}

	//／计算工 地址在网段位图数组中的索引位置
	c := 0
	//将 IP 地址转换成 4个字节的表示方式
	releaseIP := ipaddr.To4()
	//／由于 IP 是从 1开始分配的，所以转换成索引应减1
	releaseIP[3]-=1
	for t := uint(4); t > 0; t-=1 {
		//＊与分配 IP 相反，释放 IP 获得索引的方式是 IP 地址的每 位相减之后分别左移将对应的数值加到索引上。
		c += int(releaseIP[t-1] - subnet.IP[t-1]) << ((4-t) * 8)
	}

	//／将分配的位图数组中索号｜位置的值置为0
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	//／保存释放掉 IP 之后的网段 IP 分配信息
	ipam.dump()
	return nil
}