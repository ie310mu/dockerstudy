## 由来
根据《自己动手写Docker.陈显鹭》学习了docker的原理，但由于linux内核的更新，其中有些代码无法正常运行，做了调整

## 开发环境
建议使用一个干净的linux虚拟机作为开发环境，避免污染系统  
安装go1.7  
需要支持aufs文件系统，如果不支持要安装插件  
每个目录都对应原书中的章节，如 /3.1/mydocker  
在/3.1/mydocker下，运行go build .进行编译  
busybox.tar放到/root/busybox.tar  