package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"io/ioutil"
	"net/http"

	"github.com/go-co-op/gocron"
	"github.com/go-ini/ini"
)

// 前言：
// 	  今天把程序重构一下，优化执行过程。
// 去除分段执行，所有步骤全部写在主分区。
// ......
// 另外用纯Golang获取ip那地方真的很难搞...

// 全局变量（配置模板）
// [dyn] # dynv6.com 的 Api密钥 和 DDNS域名
var token, domain string = "1234567890AaBbCc", "example.dynv6.net"

// [ip] # 获取到的本地ip，用于检测是否变化，填disable禁用
var v4, v6 string = "127.0.0.1", "::1"

// [cron] # 计划任务执行间隔时间，单位 秒，和 是否显示ip变化
var wait, show string = "600", "false"

// 初始化配置文件
func reconf() {
	fmt.Println("\n初始化配置文件...")
	// 调用ini包写入默认配置
	cfg := ini.Empty()
	cfg.Section("dyn").NewKey("token", token)
	cfg.Section("dyn").NewKey("domain", domain)
	cfg.Section("ip").NewKey("v4", v4)
	cfg.Section("ip").NewKey("v6", v6)
	cfg.Section("cron").NewKey("wait", wait)
	cfg.Section("cron").NewKey("show", show)
	cfg.SaveTo("conf.ini")
	// 提示并退出程序
	fmt.Println("初始化成功，请打开conf.ini进行配置！")
	os.Exit(0)
	//rexit()
}

// win端按任意键退出（有Bug，已弃用）
/*func rexit() {
	fmt.Printf("\n> 按回车键退出...")
	b := make([]byte, 1)
	os.Stdin.Read(b)
	os.Exit(0)
}*/

// 主程序
func main() {
	fmt.Print(`
--------------------------------------
| EASYDYNV6 - A DDNS PROGRAM BY ZXWY |
|  ORIG: https://dynv6.com VER: 1.2  |
--------------------------------------
`)
	// 1. 检查配置文件
	fmt.Println("\n> 检查配置文件")
	// 1.1 是否有配置文件
	cfg, err := ini.Load("conf.ini")
	if err != nil {
		// 如果没有读取到配置文件，则创建
		reconf()
	}
	// 1.11 检测配置项是否完整
	y := cfg.Section("dyn").HasKey("token")
	if y == false {
		fmt.Println("\n关键项目丢失，重写为默认配置。")
		reconf()
	}
	y2 := cfg.Section("dyn").HasKey("domain")
	if y2 == false {
		fmt.Println("\n关键项目丢失，重写为默认配置。")
		reconf()
	}
	// 1.2 将配置文件读入变量
	dyntoken := cfg.Section("dyn").Key("token").String()
	dyndomain := cfg.Section("dyn").Key("domain").String()
	// 1.3 检查配置（简单检测，只能判断是否修改，无法检测配置有效性）
	fmt.Println("\n检测配置内容...")
	// 如果有错误则结束程序
	var e int = 0
	if dyntoken == token {
		e++
		fmt.Println("[dyn] 请修改 token 配置！")
	} else {
		if dyntoken == "" {
			e++
			fmt.Println("[dyn] 请填写 token 数据！")
		}
	}
	if dyndomain == domain {
		e++
		fmt.Println("[dyn] 请修改 domain 配置！")
	} else {
		if dyndomain == "" {
			e++
			fmt.Println("[dyn] 请填写 domain 数据！")
		}
	}
	cronwait, err := cfg.Section("cron").Key("wait").Int()
	if err != nil {
		e++
		fmt.Println("[cron] wait 值仅可填入数字！")
	} else {
		if cronwait <= 0 {
			e++
			fmt.Println("[cron] wait 值需大于 0 ！")
		}
	}
	cronshow, err := cfg.Section("cron").Key("show").Bool()
	if err != nil {
		e++
		fmt.Println("[cron] show 值仅可填入true或false！")
	}
	if e > 0 {
		//fmt.Println("[info] 如不小心清空配置文件，请删除后重新生成。")
		fmt.Println("\n检测不通过，共", e, "处错误，请查看输出信息。")
		os.Exit(1)
		//rexit()
	}
	fmt.Print("检测通过！\n\n> 执行计划任务\n\n")
	// 2. 运行计划任务
	i := 0
	s := gocron.NewScheduler(time.Local)
	s.Every(cronwait).Seconds().Do(func() {
		// 循环执行以下DDNS脚本
		i++
		log.Println("第", i, "次执行")
		// 2.1 读取配置ip
		cfg, _ := ini.Load("conf.ini") // 实现动态加载
		ipv4 := cfg.Section("ip").Key("v4").String()
		ipv6 := cfg.Section("ip").Key("v6").String()
		// 2.2 读取本地ip
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			fmt.Println("[ddns] 本地ip读取失败：", err)
		}
		// 是否使用ipv4
		if ipv4 == "disable" {
			fmt.Println("[ddns] 已禁用ipv4数据更新")
		} else {
			for _, addr := range addrs {
				ipNet, isIpNet := addr.(*net.IPNet)
				if isIpNet && ipNet.IP.IsGlobalUnicast() {
					// ipv4操作
					if ipNet.IP.To4() != nil {
						gipv4 := ipNet.IP.String()
						if cronshow == true {
							fmt.Println("[show] 配置ipv4:", ipv4)
							fmt.Println("[show] 本地ipv4:", gipv4)
						}
						// 更新ip地址
						if ipv4 == gipv4 {
							fmt.Println("[ipv4] ip未发生变化，跳过更新")
						} else {
							// 将新ip地址写入配置
							cfg.Section("ip").Key("v4").SetValue(gipv4)
							cfg.SaveTo("conf.ini")
							fmt.Println("[ipv4] ip发生变化，写入配置文件")
							url4 := "http://ipv4.dynv6.com/api/update?hostname=" + dyndomain + "&ipv4=" + gipv4 + "&token=" + dyntoken
							resp, err := http.Get(url4)
							if err != nil {
								fmt.Println(err)
								break
							}
							defer resp.Body.Close()
							body, _ := ioutil.ReadAll(resp.Body)
							//fmt.Println("[ipv4] 提交至Api接口，返回", string(body))
							if string(body) == "invalid authentication token" {
								fmt.Println("[err] 更新失败，Api token 错误！")
								// 写为默认配置，防止再次执行
								cfg.Section("dyn").Key("token").SetValue(token)
								cfg.Section("ip").Key("v4").SetValue(v4)
								cfg.SaveTo("conf.ini")
								//rexit()
								os.Exit(1)
							} else {
								if string(body) == "zone not found" {
									fmt.Println("[err] 更新失败，DDNS 域名错误！")
									// 写为默认配置，防止再次执行
									cfg.Section("dyn").Key("domain").SetValue(domain)
									cfg.Section("ip").Key("v4").SetValue(v4)
									cfg.SaveTo("conf.ini")
									//rexit()
									os.Exit(1)
								} else {
									if string(body) == "addresses updated" {
										fmt.Println("[ipv4] 更新成功！")
									} else {
										fmt.Println("[ipv4] 服务端地址未改变！")
									}
								}
							}
						}
						break
					}
				}
			}
		}
		// 是否使用ipv6
		if ipv6 == "disable" {
			fmt.Println("[ddns] 已禁用ipv6数据更新")
		} else {
			for _, addr := range addrs {
				ipNet, isIpNet := addr.(*net.IPNet)
				if isIpNet && ipNet.IP.IsGlobalUnicast() {
					// ipv6操作
					if ipNet.IP.To4() == nil {
						gipv6 := ipNet.IP.String()
						if cronshow == true {
							fmt.Println("[show] 配置ipv6:", ipv6)
							fmt.Println("[show] 本地ipv6:", gipv6)
						}
						// 更新ip地址
						if ipv6 == gipv6 {
							fmt.Println("[ipv6] ip未发生变化，跳过更新")
						} else {
							// 将新ip地址写入配置
							cfg.Section("ip").Key("v6").SetValue(gipv6)
							cfg.SaveTo("conf.ini")
							fmt.Println("[ipv6] ip发生变化，写入配置文件")
							url6 := "http://dynv6.com/api/update?hostname=" + dyndomain + "&ipv6=" + gipv6 + "&token=" + dyntoken
							resp, err := http.Get(url6)
							if err != nil {
								fmt.Println(err)
								break
							}
							defer resp.Body.Close()
							body, _ := ioutil.ReadAll(resp.Body)
							//fmt.Println("[ipv6] 提交至Api接口，返回", string(body))
							if string(body) == "invalid authentication token" {
								fmt.Println("[err] 更新失败，Api token 错误！")
								// 写为默认配置，防止再次执行
								cfg.Section("dyn").Key("token").SetValue(token)
								cfg.SaveTo("conf.ini")
								//rexit()
								os.Exit(1)
							} else {
								if string(body) == "zone not found" {
									fmt.Println("[err] 更新失败，DDNS 域名错误！")
									// 写为默认配置，防止再次执行
									cfg.Section("dyn").Key("domain").SetValue(domain)
									cfg.SaveTo("conf.ini")
									//rexit()
									os.Exit(1)
								} else {
									if string(body) == "addresses updated" {
										fmt.Println("[ipv6] 更新成功！")
									} else {
										fmt.Println("[ipv6] 服务端地址未改变！")
									}
								}
							}
						}
						break
					}
				}
			}
		}
		// 2. 等待下次执行
		fmt.Println()
	})
	s.StartBlocking()
}
