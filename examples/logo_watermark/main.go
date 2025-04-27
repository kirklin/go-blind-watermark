package main

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/kirklin/go-blind-watermark/bwm"
)

func main() {
	// 1. 读取宿主图 (Host Image)
	hostFile, _ := os.Open("test.png")
	defer hostFile.Close()
	hostImg, _, _ := image.Decode(hostFile)

	// 2. 读取 Logo 图 (Watermark Image)
	logoFile, _ := os.Open("logo.png")
	defer logoFile.Close()
	logoImg, _, _ := image.Decode(logoFile)

	// 3. 处理 Logo
	// Logo 不能太大，因为宿主图容量有限。
	// 容量估算：(HostW / 4) * (HostH / 4) 是最大比特数。
	// 比如 1920x1080 的图，大约能存 480*270 = 129,600 bits
	// 也就是 Logo 最好缩放到 64x64 或者 128x128 这种级别。

	targetW, targetH := 64, 64
	resizedLogo := bwm.ResizeImage(logoImg, targetW, targetH)

	// 转比特 (阈值 128)
	wmBits, w, h := bwm.LogoToBits(resizedLogo, 128)
	fmt.Printf("Logo 比特流大小: %dx%d (%d bits)\n", w, h, len(wmBits))

	// 4. 嵌入
	engine := bwm.New(999, 888)
	engine.D1 = 36.0 // 默认强度

	outputImg, _ := engine.Embed(hostImg, wmBits)

	// 5. 保存
	outF, _ := os.Create("test_with_logo.png")
	defer outF.Close()
	png.Encode(outF, outputImg)
	fmt.Println("Logo 嵌入完成。")

	// --- 提取演示 ---
	// 提取时，你需要知道当时嵌入的 Logo 尺寸 (64x64)
	fmt.Println("正在提取 Logo...")

	exBits, _ := engine.Extract(outputImg, w*h)

	// 还原为图片
	recoveredLogo := bwm.BitsToLogo(exBits, w, h)

	saveLogo, _ := os.Create("recovered_logo.png")
	png.Encode(saveLogo, recoveredLogo)
	saveLogo.Close()
	fmt.Println("Logo 提取并保存为 recovered_logo.png")
}
