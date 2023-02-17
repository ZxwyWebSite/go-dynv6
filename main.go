package main

// 说明：第一次自己从头写Golang程序，可能包含大量重复代码、逻辑混乱、等，
// 以后会改的，就问你程序能不能跑，能跑就别动，动了就不一定能跑起来了。

// 引入包(如缺包请执行 go install {包地址}@latest 😅)
import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-ini/ini"
)

// 欢迎页面
func main() {
	fmt.Print(`
     _                        _   
    | |                     / /   
 ___| | _  __ _____  _  _  / /__  
/  _  |\ \/ //  _  \\ \/ //  _  \ 
| |_| | \  / | | | | \  / | |_| | 
\___/_| / /  |_| |_|  \/  \_____/ 
       /_/   dynv6.com     DDNS   
==================================
 Ver 1.0  Build 20230217  By Zxwy 
`)
	ckconf()
	//task()
	//test()
}

// 检查配置文件
func ckconf() {
	fmt.Println("\n> 检查配置文件")
	// 读取配置文件
	cfg, err := ini.Load("conf.ini")
	if err != nil {
		fmt.Println("创建配置文件...")
		crconf()
	}
	// 将参数读入变量
	dyntoken := cfg.Section("dyn").Key("token").In("err", []string{"1234567890"})
	dyndomain := cfg.Section("dyn").Key("domain").In("err", []string{"example.dynv6.net"})
	//ipv4 := cfg.Section("ip").Key("v4").String()
	//ipv6 := cfg.Section("ip").Key("v6").String()
	// 简单验证更改
	if dyntoken != string("err") {
		fmt.Println("\n请填写token参数！")
		os.Exit(1)
	}
	if dyndomain != string("err") {
		fmt.Println("\n请填写domain参数！")
		os.Exit(1)
	}
	fmt.Println("\n通过检查！")
	//fmt.Println("\n[dyn]", "\ntoken =", dyntoken, "\ndomain =", dyndomain, "\n\n[ip]", "\nv4 =", ipv4, "\nv6 =", ipv6)
	fmt.Println("\n> 执行计划任务")
	cron()
}

// 创建配置文件
func crconf() {
	file, err := os.Create("conf.ini")
	if err != nil {
		fmt.Println("创建失败，请检查用户权限！", err)
		os.Exit(1)
	}
	// 配置文件模板
	str := "[dyn]\ntoken = 1234567890\ndomain = example.dynv6.net\n\n[ip]\nv4 = 127.0.0.1\nv6 = ::1"
	file.Write([]byte(str)) //将字符串转换成字节切片
	fmt.Println("创建成功，请打开conf.ini进行配置")
	defer file.Close()
	os.Exit(1)
}

// 配置计划任务
func cron() {
	i := 0
	s := gocron.NewScheduler(time.Local)
	// 默认每分钟执行一次，之后可能会允许在配置中修改
	s.Every(1).Minutes().Do(func() {
		i++
		fmt.Println()
		log.Println("第", i, "次执行")
		task()
	})
	s.StartBlocking()
}

// 获取本地ip4地址
func getip4() (gipv4 string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
	}
	for _, addr := range addrs {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && ipNet.IP.IsGlobalUnicast() {
			if ipNet.IP.To4() != nil {
				gipv4 := ipNet.IP.String()
				//fmt.Println(gipv4)
				return gipv4
			}
		}
	}
	return
}

// 获取本地ip6地址
func getip6() (gipv6 string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
	}
	for _, addr := range addrs {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && ipNet.IP.IsGlobalUnicast() {
			if ipNet.IP.To4() == nil {
				gipv6 := ipNet.IP.String()
				//fmt.Println(gipv6)
				return gipv6
			}
		}
	}
	return
}

// 重构版DDNS脚本
func task() {
	// 初始化局部变量
	cfg, _ := ini.Load("conf.ini")
	dyntoken := cfg.Section("dyn").Key("token").String()
	dyndomain := cfg.Section("dyn").Key("domain").String()
	ipv4 := cfg.Section("ip").Key("v4").String()
	ipv6 := cfg.Section("ip").Key("v6").String()
	gipv4 := getip4()
	gipv6 := getip6()
	// 对比配置文件
	if ipv4 == "disable" {
		fmt.Println("[v4] 已禁用ipv4 DDNS")
	} else {
		if ipv4 == gipv4 {
			fmt.Println("[v4] ip未发生变化，跳过更新")
		} else {
			cfg.Section("ip").Key("v4").SetValue(gipv4)
			cfg.SaveTo("conf.ini")
			fmt.Println("[v4] ip发生变化，已写入配置文件")
			url4 := "http://ipv4.dynv6.com/api/update?hostname=" + dyndomain + "&ipv4=" + gipv4 + "&token=" + dyntoken
			resp, err := http.Get(url4)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
		}
	}
	if ipv6 == "disable" {
		fmt.Println("[v6] 已禁用ipv6 DDNS")
	} else {
		if ipv6 == gipv6 {
			fmt.Println("[v6] ip未发生变化，跳过更新")
		} else {
			cfg.Section("ip").Key("v6").SetValue(gipv6)
			cfg.SaveTo("conf.ini")
			fmt.Println("[v6] ip发生变化，已写入配置文件")
			url6 := "http://dynv6.com/api/update?hostname=" + dyndomain + "&ipv6=" + gipv6 + "&token=" + dyntoken
			resp, err := http.Get(url6)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer resp.Body.Close()
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
		}
	}

}
