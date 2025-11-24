package main

import (
	"flag"
	"log"
	"os"
)

func main() {
	configPath := flag.String("config", "migrate-config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	config, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 创建迁移器
	migrator, err := NewMigrator(config)
	if err != nil {
		log.Fatalf("创建迁移器失败: %v", err)
	}

	// 执行迁移
	if err := migrator.MigrateAll(); err != nil {
		log.Fatalf("迁移失败: %v", err)
	}

	log.Println("所有数据迁移完成！")
	os.Exit(0)
}

