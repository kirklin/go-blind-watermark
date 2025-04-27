// Copyright (c) 2025 Kirk Lin
// SPDX-License-Identifier: MIT

package bwm

import (
	"image"
	"image/color"
	"math/rand"

	"gonum.org/v1/gonum/mat"
)

// YUVData 存储图像的 YUV 分量矩阵
type YUVData struct {
	Y *mat.Dense
	U *mat.Dense
	V *mat.Dense
	W int
	H int
}

// ImageToYUV 将 image.Image 转换为 YUV 矩阵
func ImageToYUV(img image.Image) *YUVData {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// 确保尺寸是偶数，方便 DWT
	validW, validH := w-(w%2), h-(h%2)

	yMat := mat.NewDense(validH, validW, nil)
	uMat := mat.NewDense(validH, validW, nil)
	vMat := mat.NewDense(validH, validW, nil)

	for j := 0; j < validH; j++ {
		for i := 0; i < validW; i++ {
			r, g, b, _ := img.At(bounds.Min.X+i, bounds.Min.Y+j).RGBA()
			// image.RGBA 返回的是 0-65535，需转为 0-255
			rf, gf, bf := float64(r>>8), float64(g>>8), float64(b>>8)

			// 标准 RGB -> YUV 转换公式
			y := 0.299*rf + 0.587*gf + 0.114*bf
			u := -0.14713*rf - 0.28886*gf + 0.436*bf
			v := 0.615*rf - 0.51499*gf - 0.10001*bf

			yMat.Set(j, i, y)
			uMat.Set(j, i, u)
			vMat.Set(j, i, v)
		}
	}

	return &YUVData{Y: yMat, U: uMat, V: vMat, W: validW, H: validH}
}

// YUVToImage 将 YUV 矩阵转回 image.Image
func YUVToImage(data *YUVData) image.Image {
	rect := image.Rect(0, 0, data.W, data.H)
	img := image.NewRGBA(rect)

	for j := 0; j < data.H; j++ {
		for i := 0; i < data.W; i++ {
			y := data.Y.At(j, i)
			u := data.U.At(j, i)
			v := data.V.At(j, i)

			r := y + 1.13983*v
			g := y - 0.39465*u - 0.58060*v
			b := y + 2.03211*u

			// Clamp results
			img.Set(i, j, color.RGBA{
				R: uint8(clamp(r)),
				G: uint8(clamp(g)),
				B: uint8(clamp(b)),
				A: 255,
			})
		}
	}
	return img
}

func clamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// GenerateShuffledIndices 生成基于 Seed 的乱序索引
func GenerateShuffledIndices(length int, seed int64) []int {
	indices := make([]int, length)
	for i := 0; i < length; i++ {
		indices[i] = i
	}
	rng := rand.New(rand.NewSource(seed))
	rng.Shuffle(length, func(i, j int) {
		indices[i], indices[j] = indices[j], indices[i]
	})
	return indices
}
