// Copyright (c) 2025 Kirk Lin
// SPDX-License-Identifier: MIT

package bwm

import (
	"image"
)

// BlindWatermark 盲水印处理器
type BlindWatermark struct {
	SeedImg int64   // 图像块乱序种子
	SeedWm  int64   // 水印位乱序种子
	D1      float64 // 嵌入强度系数 1 (Coarse)
	D2      float64 // 嵌入强度系数 2 (Fine, unused in basic mode)
}

// New 创建一个新的处理器
func New(seedImg, seedWm int64) *BlindWatermark {
	return &BlindWatermark{
		SeedImg: seedImg,
		SeedWm:  seedWm,
		D1:      36.0, // 经验值，可调整
		D2:      20.0,
	}
}

// Embed 将水印比特流嵌入图片
func (b *BlindWatermark) Embed(src image.Image, wmBits []bool) (image.Image, error) {
	// 1. 转 YUV
	yuv := ImageToYUV(src)

	// 2. 执行嵌入核心流程
	resYUV := b.EmbedProcess(yuv, wmBits)

	// 3. 转回 Image
	return YUVToImage(resYUV), nil
}

// Extract 从图片中提取水印比特流
func (b *BlindWatermark) Extract(src image.Image, wmLength int) ([]bool, error) {
	// 1. 转 YUV
	yuv := ImageToYUV(src)

	// 2. 执行提取核心流程
	return b.ExtractProcess(yuv, wmLength), nil
}
