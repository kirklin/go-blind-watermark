package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/kirklin/go-blind-watermark/bwm"
)

func main() {
	// --- 1. 加载原图 ---
	srcFile, err := os.Open("test.png")
	if err != nil {
		panic(err)
	}
	defer srcFile.Close()
	srcImg, _, err := image.Decode(srcFile)
	if err != nil {
		panic(err)
	}

	// --- 2. 准备水印内容 ---
	text := "Copyright 2025 Kirk Lin - All Rights Reserved"
	fmt.Printf("水印文字: %s\n", text)

	// 转换为比特流
	wmBits := bwm.TextToBits(text)
	fmt.Printf("水印比特长度: %d bits\n", len(wmBits))

	// --- 3. 嵌入水印 ---
	// 密钥种子: 12345 (图片混淆), 67890 (水印加密)
	engine := bwm.New(12345, 67890)

	// 设置强度 (D1越大越抗攻击，但画质越差)
	engine.D1 = 40.0

	watermarkedImg, err := engine.Embed(srcImg, wmBits)
	if err != nil {
		panic(err)
	}

	// --- 4. 保存结果 ---
	outFile, _ := os.Create("test_with_text.png")
	defer outFile.Close()
	png.Encode(outFile, watermarkedImg)
	fmt.Println("加密完成: test_with_text.png")

	// ==========================================
	// 模拟提取测试
	// ==========================================

	fmt.Println("\n--- 正在尝试提取 ---")

	// 假设我们只知道水印的大概长度，或者即使读取多一点也没关系
	// 英文每个字符 8 bit，我们提取同样的长度
	extractLen := len(wmBits)

	extractedBits, _ := engine.Extract(watermarkedImg, extractLen)

	extractedText := bwm.BitsToText(extractedBits)
	fmt.Printf("提取结果: %s\n", extractedText)
}
