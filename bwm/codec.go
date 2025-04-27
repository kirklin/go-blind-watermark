// Copyright (c) 2025 Kirk Lin
// SPDX-License-Identifier: MIT

package bwm

import (
	"image"
	"image/color"
)

// --- 文本编解码 ---

// TextToBits 将字符串转换为比特数组
// 采用 UTF-8 编码，每个字节转为 8 位
func TextToBits(text string) []bool {
	bytes := []byte(text)
	bits := make([]bool, 0, len(bytes)*8)

	for _, b := range bytes {
		for i := 7; i >= 0; i-- {
			// 检查每一位是否为 1
			bit := (b>>i)&1 == 1
			bits = append(bits, bit)
		}
	}
	return bits
}

// BitsToText 将比特数组还原为字符串
// 自动丢弃末尾不足 8 位的部分
func BitsToText(bits []bool) string {
	var bytes []byte
	numBytes := len(bits) / 8

	for i := 0; i < numBytes; i++ {
		var b byte
		for j := 0; j < 8; j++ {
			if bits[i*8+j] {
				b |= 1 << (7 - j)
			}
		}
		bytes = append(bytes, b)
	}
	return string(bytes)
}

// --- 图片（Logo）编解码 ---

// LogoToBits 将水印图片二值化并转为比特数组
// threshold: 二值化阈值 (0-255)，通常取 128
func LogoToBits(img image.Image, threshold uint8) ([]bool, int, int) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	bits := make([]bool, 0, w*h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// 获取灰度值
			c := img.At(bounds.Min.X+x, bounds.Min.Y+y)
			gray := color.GrayModel.Convert(c).(color.Gray).Y

			// 大于阈值视为 1 (白/亮)，小于视为 0 (黑/暗)
			// 注意：通常水印是黑色字，白色底。我们把黑色(暗)作为信息位(true)更符合直觉，
			// 但这里标准做法是：位图里 1 是数据。我们定义：亮色=false, 深色=true (有墨水)
			isInk := gray < threshold
			bits = append(bits, isInk)
		}
	}
	return bits, w, h
}

// BitsToLogo 将提取出的比特流还原为黑白图片
func BitsToLogo(bits []bool, w, h int) image.Image {
	rect := image.Rect(0, 0, w, h)
	img := image.NewGray(rect)

	// 限制长度防止越界
	limit := len(bits)
	if limit > w*h {
		limit = w * h
	}

	for i := 0; i < limit; i++ {
		x := i % w
		y := i / w

		val := uint8(255) // 默认为白
		if bits[i] {
			val = 0 // 如果是 bit 1，则还原为黑 (墨水)
		}

		img.SetGray(x, y, color.Gray{Y: val})
	}
	return img
}

// ResizeImage 简单的图片缩放工具 (Nearest Neighbor)，用于调整 Logo 大小以适应载体
func ResizeImage(img image.Image, newW, newH int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	srcBounds := img.Bounds()
	srcW, srcH := srcBounds.Dx(), srcBounds.Dy()

	xRatio := float64(srcW) / float64(newW)
	yRatio := float64(srcH) / float64(newH)

	for y := 0; y < newH; y++ {
		for x := 0; x < newW; x++ {
			srcX := int(float64(x) * xRatio)
			srcY := int(float64(y) * yRatio)
			dst.Set(x, y, img.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY))
		}
	}
	return dst
}
