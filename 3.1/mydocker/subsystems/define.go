// package subsystems

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"os"
// 	"path"
// 	"strconv"
// 	"bufio"
// 	"strings"
// 	"github.com/Sirupsen/logrus"
// )

// //用于传递资源限制配直的结构体，包含内存限制， CPU 时间片权重 CPU 核心数
// type ResourceConfig struct{
// 	MemoryLimit string  //内存限制
// 	CpuShare string  //CPU 时间片权重
// 	CpuSet string	 //CPU 核心数
// }

// //Subsystem 接口，每个 Subsystem 可以实现下面的 个接口
// //／这里将 cgroup 抽象成了 path 原因是 cgroup hierarchy 的路径，便是虚拟文件系统中的虚拟路径
// type SubSystem interface{
// 	Name() string   //返回 subsystem 的名字，比如 cpu memory . 
// 	Set(path string, res *ResourceConfig) error //设置某个 cgroup 在这个 Subsystem 中的资源限制
// 	Apply(path string, pid int) error //将迸程添加到某个 cgroup中
// 	Remove(path string)  error //移除某个 cgroup
// }

// //通过不同 subsystem 初始化实例创建资源限制处理链数组

// var (
// 	SubSystemIns = []SubSystem{
// 		// &CpusetSubSystem{},
// 		&MemorySubSystem{},
// 		// &CpuSubSystem{},
// 	}
// )

// // memory subsystem 的实现
// type MemorySubSystem struct{

// }

// //设置 cgroupPath 对应的 cgroup 的内存资源限制
// func (s *MemorySubSystem) Set(cgroupPath string , res *ResourceConfig) error {
// 	//GetCgroupPath 的作用是获取当前 subsystem 在虚拟文件系统中的路径， GetCgroupPath 这个函数在下面会介绍。
// 	if subsysCgroupPath, err := GetCgroupPath(s.Name() , cgroupPath, true); err== nil {
// 		if res.MemoryLimit != "" {
// 			//设置这个 cgroup 的内存限制，即将限制写入到 cgroup 对应目录的 memory.limit in bytes文件中。
// 			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes") , []byte(res.MemoryLimit) , 0644); err != nil {
// 				return fmt.Errorf("set cgroup memory fail %v", err)
// 			}
// 		} 
// 		return nil
// 	}else {
// 		return err
// 	}
// }

// //删除 cgroupPath 对应的 cgroup
// func (s *MemorySubSystem) Remove(cgroupPath string) error {
// 	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil{
// 		//删除 cgroup 便是删除对应的 cgroupPath 的目录
// 		return os.Remove(subsysCgroupPath)
// 	}else{
// 		return err
// 	}
// }

// //将一个进程加入到 cgroupPath 对应的 cgroup中
// func (s *MemorySubSystem) Apply(cgroupPath string, pid int) error{
// 	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil{
// 		//把进程的 PID 写到 cgroup 的虚拟文件系统对应目录下的” task ”文件中
// 		if err := 	ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil{
// 			return fmt.Errorf("set cgroup proc fail %v", err)
// 		}
// 		return nil
// 	}else{
// 		return fmt.Errorf("get cgroup %s error: %v",cgroupPath, err)
// 	}
// }

// //返回 cgroup 的名字
// func (s *MemorySubSystem) Name() string{
// 	return "memory"
// }

// //通过／proc/self/mountinfo 找出挂载了某个 subsystem hierarchy cgroup 根节点所在的目录FindCgroupMountpoint (” memory” ) 
// func FindCgroupMountpoint(subsystem string) string{
// 	f, err := os.Open("/proc/self/mountinfo")
// 	if err != nil{
// 		return ""
// 	}
// 	defer f.Close()

// 	scanner := bufio.NewScanner(f)
// 	for scanner.Scan() {
// 		txt := scanner.Text()
// 		fields := strings.Split(txt, " ")
// 		for _, opt := range strings.Split(fields[len(fields) - 1], ","){
// 			if opt == subsystem{
// 				return fields[4]
// 			}
// 		}
// 	}
// 	if err := scanner.Err(); err != nil{
// 		return "";
// 	}
// 	return "";
// }


// //／得到 cgroup 在文件系统中的绝对路径
// func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error){
// 	cgroupRoot := FindCgroupMountpoint(subsystem)
// 	if _, err := os.Stat(path.Join(cgroupRoot, cgroupPath)); err == nil || (autoCreate && os.IsNotExist(err)){
// 		if os.IsNotExist(err) {
// 			if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err == nil{				
// 			}else{
// 				return "", fmt.Errorf("error create cgroup %v", err)
// 			}
// 		}
// 	}else{
// 		return "", fmt.Errorf("cgroup path error %v", err)
// 	}

// 	return path.Join(cgroupRoot, cgroupPath), nil
// }


// type CgroupManager struct{
// 	//cgroup hierarchy 中的路径 相当于创建的 cgroup 目录相对于各 root cgroup 目录的路径
// 	Path string
// 	Resource *ResourceConfig
// }

// func NewCgroupManager(path string) *CgroupManager{
// 	return &CgroupManager{
// 		Path: path,
// 	}
// }

// //将进程 PID 加入到每个 cgroup中
// func (c *CgroupManager) Apply(pid int) error{
// 	for _, subSysIns := range(SubSystemIns){
// 		subSysIns.Apply(c.Path, pid)
// 	}
// 	return nil
// }

// //设置各个 subsystem 挂载中的 cgroup 资源限制
// func (c *CgroupManager) Set(res *ResourceConfig) error{
// 	for _, subSysIns := range(SubSystemIns){
// 		subSysIns.Set(c.Path, res)
// 	}
// 	return nil
// }

// //释放各个 subsystem 挂裁中的 cgroup
// func (c *CgroupManager) Destory() error{
// 	for _, subSysIns := range(SubSystemIns){
// 		if err := subSysIns.Remove(c.Path); err != nil{
// 			logrus.Warnf("remove cgroup fail %v", err)
// 		}
// 	}
// 	return nil
// }

