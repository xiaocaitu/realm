package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

type ForwardingRule struct {
	Listen string `toml:"listen"`
	Remote string `toml:"remote"`
}

type Config struct {
	Network struct {
		NoTCP  bool `toml:"no_tcp"`
		UseUDP bool `toml:"use_udp"`
	} `toml:"network"`
	Endpoints []ForwardingRule `toml:"endpoints"`
}

type PanelConfig struct {
	Auth struct {
		Password string `toml:"password"`
	} `toml:"auth"`
	Server struct {
		Port int `toml:"port"`
	} `toml:"server"`
	HTTPS struct {
		Enabled  bool   `toml:"enabled"`
		CertFile string `toml:"cert_file"`
		KeyFile  string `toml:"key_file"`
	} `toml:"https"`
}

var (
	rules       []ForwardingRule
	mu          sync.Mutex
	config      Config
	panelConfig PanelConfig
)

func LoadConfig() error {
	data, err := ioutil.ReadFile("/root/.realm/config.toml")
	if err != nil {
		return err
	}

	if _, err := toml.Decode(string(data), &config); err != nil {
		return err
	}

	rules = config.Endpoints
	return nil
}

func LoadPanelConfig() error {
	data, err := ioutil.ReadFile("./config.toml")
	if err != nil {
		return err
	}

	if _, err := toml.Decode(string(data), &panelConfig); err != nil {
		return err
	}

	return nil
}

func SaveRules() error {
	mu.Lock()
	defer mu.Unlock()

	config.Endpoints = rules
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	return ioutil.WriteFile("/root/.realm/config.toml", data, 0644)
}

// 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user == nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	if err := LoadConfig(); err != nil {
		log.Fatalf("无法加载 realm 配置: %v", err)
	}

	if err := LoadPanelConfig(); err != nil {
		log.Fatalf("无法加载面板配置: %v", err)
	}

	r := gin.Default()

	// 设置 session
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("realm_session", store))

	// 静态文件
	r.Static("/static", "./static")

	// 登录页面
	r.GET("/login", func(c *gin.Context) {
		session := sessions.Default(c)
		if session.Get("user") != nil {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.File("./templates/login.html")
	})

	// 登录处理
	r.POST("/login", func(c *gin.Context) {
		var loginData struct {
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&loginData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求"})
			return
		}

		if loginData.Password == panelConfig.Auth.Password {
			session := sessions.Default(c)
			session.Set("user", true)
			session.Options(sessions.Options{
				MaxAge: 3600 * 24, // 24小时
			})
			if err := session.Save(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Session保存失败"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "登录成功"})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		}
	})

	// 登出
	r.POST("/logout", AuthRequired(), func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
	})

	// 需要认证的路由
	authorized := r.Group("/")
	authorized.Use(AuthRequired())
	{
		// 主页
		authorized.GET("/", func(c *gin.Context) {
			c.File("./templates/index.html")
		})

		// 获取转发规则
		authorized.GET("/get_rules", func(c *gin.Context) {
			mu.Lock()
			defer mu.Unlock()
			c.JSON(200, rules)
		})

		// 添加转发规则
		authorized.POST("/add_rule", func(c *gin.Context) {
			var input struct {
				Listen string `json:"listen"`
				Remote string `json:"remote"`
			}

			if err := c.ShouldBindJSON(&input); err != nil {
				c.JSON(400, gin.H{"error": "Invalid input"})
				return
			}

			mu.Lock()
			rules = append(rules, ForwardingRule{
				Listen: input.Listen,
				Remote: input.Remote,
			})
			mu.Unlock()

			if err := SaveRules(); err != nil {
				c.JSON(500, gin.H{"error": "Failed to save rules"})
				return
			}

			c.JSON(201, input)
		})

		// 删除转发规则
		authorized.DELETE("/delete_rule", func(c *gin.Context) {
			listen := c.Query("listen")

			mu.Lock()
			for i, rule := range rules {
				if rule.Listen == listen {
					rules = append(rules[:i], rules[i+1:]...)
					break
				}
			}
			mu.Unlock()

			if err := SaveRules(); err != nil {
				c.JSON(500, gin.H{"error": "Failed to save rules"})
				return
			}

			c.Status(200)
		})

		// 启动服务
		authorized.POST("/start_service", func(c *gin.Context) {
			cmd := exec.Command("systemctl", "start", "realm")
			if err := cmd.Run(); err != nil {
				c.JSON(500, gin.H{"error": "Failed to start service"})
				return
			}

			c.JSON(200, gin.H{"message": "Service started successfully"})
		})

		// 停止服务
		authorized.POST("/stop_service", func(c *gin.Context) {
			cmd := exec.Command("systemctl", "stop", "realm")
			if err := cmd.Run(); err != nil {
				c.JSON(500, gin.H{"error": "Failed to stop service"})
				return
			}

			c.JSON(200, gin.H{"message": "Service stopped successfully"})
		})

		// 检查服务状态
		authorized.GET("/check_status", func(c *gin.Context) {
			cmd := exec.Command("systemctl", "is-active", "--quiet", "realm")
			err := cmd.Run()

			var status string
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					if exitError.ExitCode() == 3 {
						status = "未启用"
					} else {
						status = "未知状态"
					}
				} else {
					status = "检查失败"
				}
			} else {
				status = "启用"
			}

			c.JSON(200, gin.H{"status": status})
		})
	}

	port := panelConfig.Server.Port
	if port == 0 {
		port = 8081 // 默认端口
	}

	if panelConfig.HTTPS.Enabled {
		if panelConfig.HTTPS.CertFile == "" || panelConfig.HTTPS.KeyFile == "" {
			log.Println("警告：HTTPS 已启用，但证书或密钥文件路径未指定。将使用 HTTP 继续。")
			log.Printf("服务器正在使用 HTTP 运行，端口：%d\n", port)
			r.Run(fmt.Sprintf(":%d", port))
		} else {
			log.Printf("服务器正在使用 HTTPS 运行，端口：%d\n", port)
			r.RunTLS(fmt.Sprintf(":%d", port), panelConfig.HTTPS.CertFile, panelConfig.HTTPS.KeyFile)
		}
	} else {
		log.Println("警告：未启用 HTTPS，将使用 HTTP 继续。")
		log.Printf("服务器正在使用 HTTP 运行，端口：%d\n", port)
		r.Run(fmt.Sprintf(":%d", port))
	}
}

