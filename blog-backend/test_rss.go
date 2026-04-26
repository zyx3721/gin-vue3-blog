package main

import (
	"blog-backend/config"
	"blog-backend/db"
	"fmt"
	"log"
)

func main() {
	// 加载配置
	if err := config.LoadConfig("config/config-dev.yml"); err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 连接数据库
	if err := db.InitDB(); err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 检查 RSS 配置
	var count int64
	db.DB.Table("settings").Where("\"group\" = ?", "rss").Count(&count)
	fmt.Printf("RSS 配置数量: %d\n", count)

	if count == 0 {
		fmt.Println("插入默认 RSS 配置...")
		// 插入默认配置
		settings := []map[string]interface{}{
			{"key": "rss_enabled", "value": "1", "type": "text", "group": "rss", "label": "RSS功能启用"},
			{"key": "rss_title", "value": "", "type": "text", "group": "rss", "label": "RSS标题"},
			{"key": "rss_description", "value": "", "type": "text", "group": "rss", "label": "RSS描述"},
			{"key": "rss_author_name", "value": "", "type": "text", "group": "rss", "label": "RSS作者名称"},
			{"key": "rss_author_email", "value": "", "type": "text", "group": "rss", "label": "RSS作者邮箱"},
			{"key": "rss_item_count", "value": "20", "type": "text", "group": "rss", "label": "RSS文章数量"},
			{"key": "rss_cache_duration", "value": "60", "type": "text", "group": "rss", "label": "RSS缓存时长(分钟)"},
		}

		for _, s := range settings {
			if err := db.DB.Table("settings").Create(s).Error; err != nil {
				log.Printf("插入配置失败 %s: %v\n", s["key"], err)
			} else {
				fmt.Printf("✓ 插入配置: %s\n", s["key"])
			}
		}
	} else {
		fmt.Println("RSS 配置已存在")
		// 显示当前配置
		var settings []struct {
			Key   string
			Value string
		}
		db.DB.Table("settings").Select("key, value").Where("\"group\" = ?", "rss").Find(&settings)
		for _, s := range settings {
			fmt.Printf("  %s = %s\n", s.Key, s.Value)
		}
	}

	fmt.Println("\n测试完成！")
}
