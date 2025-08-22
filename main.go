package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Serajian/cache-db.git/database"
)

func main() {
	// مسیر دایرکتوری برای فایل‌های persist
	basePath := "./data"

	// ساخت DB با TTL پیش‌فرض 2 ثانیه
	db := database.NewDatabase[string, string](2*time.Second, basePath)

	// Set با TTL پیش‌فرض
	db.Set("foo", "bar")

	// ذخیره روی فایل
	if err := db.Persist("test.gob"); err != nil {
		log.Fatal("persist failed:", err)
	}
	fmt.Println("✅ Persisted to", basePath+"/test.gob")

	//ساخت DB جدید و Load از فایل
	db2 := database.NewDatabase[string, string](0, basePath)
	if err := db2.Load("test.gob"); err != nil {
		log.Fatal("load failed:", err)
	}

	// گرفتن مقدار بعد از Load
	if v, ok := db2.Get("foo"); ok {
		fmt.Println("✅ Loaded value:", v)
	} else {
		fmt.Println("❌ Key missing after load")
	}

	// تست Expire
	fmt.Println("⏳ waiting 3 seconds...")
	time.Sleep(3 * time.Second)

	if _, ok := db2.Get("foo"); !ok {
		fmt.Println("✅ foo expired as expected")
	} else {
		fmt.Println("❌ foo should be expired but is still alive")
	}

	// پاک‌کردن فایل
	if err := db2.DeleteFile("test.gob"); err != nil {
		log.Fatal("delete file failed:", err)
	}
	fmt.Println("🗑 Deleted", basePath+"/test.gob")
}
